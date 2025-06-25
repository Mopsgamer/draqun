import { format, increment, parse, type ReleaseType } from "@std/semver";
import { Octokit } from "@octokit/rest";
import denojson from "../deno.json" with { type: "json" };
import { logRelease } from "./tool/constants.ts";
import { existsSync, expandGlob } from "@std/fs";

const isDryRun = Deno.args.includes("--dry-run");

if (!Deno.env.has("GITHUB_TOKEN")) {
    logRelease.error("Evironemnt variable GITHUB_TOKEN is not set.");
    Deno.exit(1);
}

const manifestPath = "deno.json";
const manifestExists = existsSync(manifestPath);

if (!manifestExists) {
    logRelease.error("Manifest file %s not found.", manifestPath);
    Deno.exit(1);
}

const newVersion = getNewVersion();
const tagName = newVersion;
const releaseName = newVersion;
const newChangelog = getNewChangelog();
logRelease.info("Increment: %s -> %s", denojson.version, newVersion);
if (isDryRun) {
    logRelease.info("Changelog:\n%s", newChangelog);
}
logRelease.info("Old latest tag: %s", getLastTag() ?? "undefined");
if (!isDryRun) {
    logRelease.start("Creating tag %s", tagName);
    createTag(tagName, newChangelog);
    logRelease.end();
    logRelease.info("New latest tag: %s", getLastTag() ?? "undefined");
}
const manifestOldContent = await Deno.readTextFile(manifestPath);
const manifestNewContent = manifestOldContent.replace(
    /("version"\s*:\s*")([^"]+)(")/,
    `$1${newVersion}$3`,
);
if (!isDryRun) {
    logRelease.start("Writing new version to %s", manifestPath);
    await Deno.writeTextFile(manifestPath, manifestNewContent);
    logRelease.end();
}
if (!isDryRun) {
    const repoInfo = getRepoInfo();
    logRelease.start("Creating GitHub release");
    await createRelease(releaseName, tagName, newChangelog, repoInfo);
    logRelease.end();
}
logRelease.success("Released successfully.");

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

    for await (
        const entry of expandGlob("dist/*", { exclude: ["dist/cache"] })
    ) {
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
    });
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
    new Deno.Command(`git`, {
        args: [
            "push",
            "origin",
            `--tags`,
        ],
    }).outputSync();
}

function getNewChangelog(from?: string): string {
    const gitLogFormat = "- %s (@%aN)%n";
    const { stdout } = new Deno.Command(`git`, {
        args: [
            "log",
            "--reverse",
            `--format=${gitLogFormat}`,
            `${from ? `${from}..` : ""}HEAD`,
        ],
    }).outputSync();

    const output = new TextDecoder().decode(stdout);
    const ignoreUsers: string[] = []; // ["dependabot[bot]", "github-actions"]
    return output.split("\n").filter((line) => {
        if (ignoreUsers.some((user) => line.endsWith(` (@${user})`))) {
            return false;
        }
        const r = /^- [a-zA-Z\d]+(\([a-zA-Z\d]+\))?!?: /g;
        return r.test(line);
    }).join("\n");
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

// deno-lint-ignore no-unused-vars
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
