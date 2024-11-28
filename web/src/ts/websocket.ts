const ws = new WebSocket("/chat/groups/1");

ws.addEventListener("message", (event) => {
    event.data;
    console.log(event.data);
});
