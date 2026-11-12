package internal

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/session"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
)

const testShardChildEnv = "DRAQUN_TEST_SHARD_CHILD"

type testShardResult struct {
	output []byte
	err    error
}

func TestMain(m *testing.M) {
	if os.Getenv(testShardChildEnv) == "1" {
		os.Exit(m.Run())
	}

	flags, canShard := shardableTestFlags(os.Args[1:])
	if !canShard {
		os.Exit(m.Run())
	}
	os.Exit(runTestShards(flags))
}

func shardableTestFlags(args []string) ([]string, bool) {
	count := "1"
	var flags []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-test.") {
			return nil, false
		}

		key, value, hasValue := strings.Cut(strings.TrimPrefix(arg, "-test."), "=")
		switch key {
		case "count":
			if !hasValue {
				if i+1 == len(args) {
					return nil, false
				}
				i++
				value = args[i]
			}
			count = value
		case "timeout":
			if !hasValue {
				if i+1 == len(args) {
					return nil, false
				}
				i++
				value = args[i]
			}
			flags = append(flags, "-test.timeout="+value)
		case "paniconexit0":
		case "testlogfile":
			if !hasValue && i+1 < len(args) {
				i++
			}
		default:
			return nil, false
		}
	}
	return append([]string{"-test.count=" + count}, flags...), true
}

func runTestShards(flags []string) int {
	// Separate processes isolate the package-global database and session state.
	list := exec.Command(os.Args[0], "-test.list=^(Test|Example)")
	list.Env = withTestShardChild(os.Environ())
	output, err := list.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list server tests: %v\n%s", err, output)
		return 1
	}

	tests := testNames(output)
	if len(tests) == 0 {
		fmt.Println("PASS")
		return 0
	}

	shardCount := min(4, runtime.GOMAXPROCS(0), len(tests))
	shards := make([][]string, shardCount)
	for i, test := range tests {
		shards[i%shardCount] = append(shards[i%shardCount], test)
	}

	results := make([]testShardResult, shardCount)
	var wg sync.WaitGroup
	for i, shard := range shards {
		wg.Add(1)
		go func(i int, shard []string) {
			defer wg.Done()
			args := append([]string{"-test.run=" + testPattern(shard)}, flags...)
			command := exec.Command(os.Args[0], args...)
			command.Env = withTestShardChild(os.Environ())
			results[i].output, results[i].err = command.CombinedOutput()
		}(i, shard)
	}
	wg.Wait()

	failed := false
	for i, result := range results {
		if result.err != nil {
			failed = true
			fmt.Fprintf(os.Stderr, "server test shard %d/%d failed: %v\n", i+1, shardCount, result.err)
		}
		if len(result.output) > 0 {
			_, _ = os.Stdout.Write(result.output)
		}
	}
	if failed {
		return 1
	}
	fmt.Println("PASS")
	return 0
}

func testNames(output []byte) []string {
	var names []string
	for _, line := range strings.Split(string(output), "\n") {
		name := strings.TrimSpace(line)
		if (strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Example")) && !strings.ContainsAny(name, " \t") {
			names = append(names, name)
		}
	}
	return names
}

func testPattern(tests []string) string {
	var pattern strings.Builder
	pattern.WriteString("^(")
	for i, test := range tests {
		if i > 0 {
			pattern.WriteByte('|')
		}
		pattern.WriteString(regexp.QuoteMeta(test))
	}
	pattern.WriteString(")$")
	return pattern.String()
}

func withTestShardChild(env []string) []string {
	for i, value := range env {
		if strings.HasPrefix(value, testShardChildEnv+"=") {
			env[i] = testShardChildEnv + "=1"
			return env
		}
	}
	return append(env, testShardChildEnv+"=1")
}

type TestServer struct {
	App      *fiber.App
	Client   *client.Client
	Listener net.Listener
}

type testListener struct {
	net.Listener
	pipeConns chan net.Conn
	tcpConns  chan net.Conn
	tcpErrors chan error
	closed    chan struct{}
	closeOnce sync.Once
}

type testConn struct {
	net.Conn
}

func (testConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

func (testConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

func newTestListener(listener net.Listener) *testListener {
	testLn := &testListener{
		Listener:  listener,
		pipeConns: make(chan net.Conn),
		tcpConns:  make(chan net.Conn),
		tcpErrors: make(chan error, 1),
		closed:    make(chan struct{}),
	}
	go testLn.acceptTCP()
	return testLn
}

func (l *testListener) acceptTCP() {
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			select {
			case l.tcpErrors <- err:
			case <-l.closed:
			}
			return
		}

		select {
		case l.tcpConns <- conn:
		case <-l.closed:
			_ = conn.Close()
			return
		}
	}
}

func (l *testListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.pipeConns:
		return testConn{Conn: conn}, nil
	case conn := <-l.tcpConns:
		return conn, nil
	case err := <-l.tcpErrors:
		return nil, err
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *testListener) DialPipe(string) (net.Conn, error) {
	clientConn, serverConn := net.Pipe()
	select {
	case l.pipeConns <- serverConn:
		return clientConn, nil
	case <-l.closed:
		_ = clientConn.Close()
		_ = serverConn.Close()
		return nil, net.ErrClosed
	}
}

func (l *testListener) Close() error {
	var err error
	l.closeOnce.Do(func() {
		close(l.closed)
		err = l.Listener.Close()
	})
	return err
}

func chdirToRoot() {
	for range 5 {
		if _, err := os.Stat("go.mod"); err == nil {
			return
		}
		_ = os.Chdir("..")
	}
}

func testSetupApp(t testing.TB) *fiber.App {
	chdirToRoot()

	// Set required environment variables
	os.Setenv("JWT_KEY", "test_jwt_key_that_is_long_enough_12345")
	os.Setenv("DB_PATH", ":memory:")

	// Close existing database if any to prevent locking issues
	if model.Sqlx != nil {
		_ = model.Sqlx.Close()
		model.Sqlx = nil
	}

	// Load environment variables
	err := environment.LoadEnv(nil)
	if err != nil {
		t.Fatalf("failed to load env: %v", err)
	}

	// Load database (automatically creates tables and runs migrations)
	err = model.LoadDB()
	if err != nil {
		t.Fatalf("failed to load db: %v", err)
	}
	db := model.Sqlx
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close test db: %v", err)
		}
		if model.Sqlx == db {
			model.Sqlx = nil
			model.Goqu = nil
		}
	})

	// Instantiate the Fiber application
	app, err := newTestApp(os.DirFS("."), false)
	if err != nil {
		t.Fatalf("failed to create fiber app: %v", err)
	}

	return app
}

func startTestServer(t testing.TB, app *fiber.App) *TestServer {
	tcpLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on TCP: %v", err)
	}
	ln := newTestListener(tcpLn)

	port := ln.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	go func() {
		_ = app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true})
	}()

	cli := client.New()
	cli.SetBaseURL(baseURL)
	cli.SetDial(ln.DialPipe)

	t.Cleanup(func() {
		cli.CloseIdleConnections()
		_ = ln.Close()
		_ = app.Shutdown()
		session.WaitForWebSocketHandlers()
	})

	return &TestServer{
		App:      app,
		Client:   cli,
		Listener: ln,
	}
}
