import { format, increment, parse, type ReleaseType } from "@std/semver";
import { Octokit } from "@octokit/rest";
import denojson from "../deno.json" with { type: "json" };
import { logRelease } from "./tool/constants.ts";
import { expandGlob } from "@std/fs";

if (!Deno.env.has("GITHUB_TOKEN")) {
    logRelease.error("Evironemnt variable missing: GITHUB_TOKEN");
    Deno.exit(1);
}

const newVersion = getNewVersion();
const newChangelog = getNewChangelog();
logRelease.info("old last tag: %s", getLastTag() ?? "undefined");
logRelease.info("new tag:\n%o", createTag(newVersion, newChangelog));
logRelease.info("new last tag: %s", getLastTag() ?? "undefined");
logRelease.warn("%s -> %s", denojson.version, newVersion);
const repoInfo = getRepoInfo();
logRelease.info("repo: %s/%s", repoInfo.owner, repoInfo.repoName);
await createRelease(newVersion, newVersion, newChangelog, repoInfo);

async function createRelease(
    name: string,
    tagName: string,
    changelog: string,
    { owner, repoName }: RepoInfo,
): Promise<void> {
    const o = new Octokit({ auth: Deno.env.get("GITHUB_TOKEN")! });
    const release = await o.repos.createRelease({
        owner,
        repo: repoName,
        tag_name: tagName,
        draft: true,

        name,
        body: changelog,
    });

    for await (const entry of expandGlob("dist/*", { exclude: ["dist/cache"] })) {
        if (!entry.isFile) continue;

        const content = await Deno.readFile(entry.path);

        await o.repos.uploadReleaseAsset({
            release_id: release.data.id,
            owner,
            repo: repoName,
            name: entry.name,
            data: new TextDecoder().decode(content),
        });
    }

    o.repos.updateRelease({
        release_id: release.data.id,
        owner,
        repo: repoName,
        draft: false,
        make_latest: "true",
    })
}

type RepoInfo = { owner: string; repoName: string };
function getRepoInfo(): RepoInfo {
    const { stdout } = new Deno.Command("git", {
        args: [
            "config",
            "--get",
            "remote.origin.url",
        ],
    }).outputSync();

    const output = new TextDecoder().decode(stdout).trim();
    const match = output.match(/github\.com[:/](.+)\/(.+?)(?:\.git)?$/);
    if (!match) {
        throw new Error(
            "Failed to parse repository information from git remote URL.",
        );
    }

    const [, owner, repoName] = match;
    return { owner, repoName };
}

function getNewVersion(): string {
    return format(increment(parse(denojson.version), getReleaseType()));
}

function createTag(name: string, changelog: string): void {
    new Deno.Command(`git`, {
        args: [
            "tag",
            "-a",
            name,
            `-m`,
            changelog,
        ],
    }).outputSync();
}

function getNewChangelog(from?: string): string {
    const gitLogFormat = "- %s%n";
    const { stdout } = new Deno.Command(`git`, {
        args: [
            "log",
            "--reverse",
            `--format=${gitLogFormat}`,
            `${from ? `${from}..` : ""}HEAD`,
        ],
    }).outputSync();

    const output = new TextDecoder().decode(stdout);
    return output;
}

function getReleaseType(): ReleaseType {
    const minor = ["feat"];
    const major = ["BREAKING CHANGE"];
    const messageList = getCommitMessages().split("\n");
    let result: ReleaseType = "patch";
    for (const message of messageList) {
        const isMajor = /^[a-zA-Z\d]+(\([a-zA-Z\d]+\))?!/g.test(message) ||
            major.some((s) => message.startsWith(s));
        if (isMajor) {
            return "major";
        }

        if (minor.some((s) => message.startsWith(s))) {
            result = "minor";
        }
    }

    return result;
}

function getLastTag(): string | undefined {
    const { code, stdout } = new Deno.Command("git", {
        args: [
            "describe",
            "--tags",
            "--match=*",
            "--no-abbrev",
            "HEAD",
        ],
    }).outputSync();

    if (code === 128) {
        return undefined;
    }

    const output = new TextDecoder().decode(stdout);
    return output.trim();
}

type Commit = {
    hash: string;
    shortHash: string;
    date: Date;
    message: string;
    body: string;
};

function getCommitMessages(from?: string): string {
    const gitLogFormat = "%s";
    const { code, stdout } = new Deno.Command(`git`, {
        args: [
            "log",
            "--reverse",
            `--format=${gitLogFormat}`,
            `${from ? `${from}..` : ""}HEAD`,
        ],
    }).outputSync();

    if (code === 128) {
        return "";
    }

    const output = new TextDecoder().decode(stdout);
    return output;
}

function getCommitLog(from?: string): Commit[] {
    const LOG_COMMIT_DELIMITER = "===LOG_COMMIT_DELIMITER===";
    const LOG_FIELD_SEPARATOR = "===LOG_FIELD_SEPARATOR===";
    const gitLogFormat =
        ["%H", "%h", "%aI", "%s", "%b"].join(LOG_FIELD_SEPARATOR) +
        LOG_COMMIT_DELIMITER;
    const { code, stdout } = new Deno.Command(`git`, {
        args: [
            "log",
            "--reverse",
            `--format=${gitLogFormat}`,
            `${from ? `${from}..` : ""}HEAD`,
        ],
    }).outputSync();

    if (code === 128) {
        return [];
    }

    const output = new TextDecoder().decode(stdout);
    const gitLog = output.split(LOG_COMMIT_DELIMITER + "\n").slice(0, -1);

    return gitLog.map((commit) => commit.split(LOG_FIELD_SEPARATOR))
        .map<Commit>((fields) => ({
            hash: fields[0],
            shortHash: fields[1],
            date: new Date(fields[2]),
            message: fields[3],
            body: fields[4],
        }));
}
