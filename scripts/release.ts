import { inc, type ReleaseType } from "semver";
import { Octokit } from "@octokit/rest";
import denojson from "../deno.json" with { type: "json" };
import { logRelease, taskDotenv } from "./tool/constants.ts";
import { existsSync, expandGlob } from "@std/fs";
import isCI from "is-ci";

const gitLogFormat = "- [%h] %s (@%cN)";
const preIdList = ["alpha", "beta"];
const preId = Deno.args.find((a) => preIdList.includes(a));

taskDotenv(logRelease);
let isDryRun = Deno.args.includes("--dry-run");

if (!isCI) {
    isDryRun = true;
    logRelease.warn("This is not an actual release.");
    logRelease.warn(
        "You are not in CI environment, dry run is enabled by default to prevent accidental releases.\n" +
            "If you really want to release, run this script in CI environment on GitHub or set 'CI' environment variable to 'true' (not recommended).",
    );
}

if (!Deno.env.has("GITHUB_TOKEN")) {
    logRelease.error("Evironemnt variable GITHUB_TOKEN is not set.");
    Deno.exit(1);
}

const manifestPath = "deno.json";
const manifestExists = existsSync(manifestPath);

if (!manifestExists) {
    logRelease.error("Manifest file" + manifestPath + "not found.");
    Deno.exit(1);
}

const latestTag = getLastTag();
const latestTagString = latestTag ?? "undefined";
const newVersion = getNewVersion();
const tagName = newVersion;
const releaseName = newVersion;
const newChangelog = getNewChangelog(latestTag);
if (isDryRun) {
    logRelease.info("Changelog:\n" + newChangelog);
}
logRelease.info("Increment: " + denojson.version + " -> " + newVersion);
logRelease.info("Old latest tag: " + latestTagString);
if (!isDryRun) {
    const task = logRelease.task({ text: "Creating tag " + tagName }).start();
    createTag(tagName, newChangelog);
    task.end("completed");
    logRelease.info("New latest tag: " + latestTagString);
}
if (newVersion !== denojson.version) {
    const manifestOldContent = await Deno.readTextFile(manifestPath);
    const manifestNewContent = manifestOldContent.replace(
        /("version"\s*:\s*")([^"]+)(")/,
        `$1${newVersion}$3`,
    );
    if (!isDryRun) {
        const task = logRelease.task({
            text: "Writing new version to " + manifestPath,
        });
        await Deno.writeTextFile(manifestPath, manifestNewContent);
        task.end("completed");
    }
}
if (!isDryRun) {
    const repoInfo = getRepoInfo();
    const task = logRelease.task({ text: "Creating GitHub release" }).start();
    await createRelease(releaseName, tagName, newChangelog, repoInfo);
    task.end("completed");
    logRelease.success("Released successfully.");
}

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
        prerelease: tagName.includes("-"),

        name,
        body: changelog,
    });

    for await (
        const entry of expandGlob("dist/server-*")
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
        make_latest: preId !== undefined ? "false" : "true",
        prerelease: preId !== undefined,
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
    let releaseType: ReleaseType;

    const found = (["major", "minor", "patch", "release"] as ReleaseType[])
        .find((a) => Deno.args.includes(a));
    let result: string | null;
    if (preId !== undefined) {
        releaseType = found ? ("pre" + found) as ReleaseType : "premajor";
        result = inc(denojson.version, releaseType, preId);
    } else {
        releaseType = found || getReleaseTypeFromCommits();
        result = inc(denojson.version, releaseType);
    }

    if (!result) throw new Error("Invalid inc result");
    return result;
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
    // if (!latestTag) {
    //     return "Initial release.";
    // }
    const { stdout } = new Deno.Command(`git`, {
        args: [
            "log",
            "--all",
            "--reverse",
            `--format=${gitLogFormat}`,
            `${from ? `${from}..` : ""}HEAD`,
        ],
    }).outputSync();

    const output = new TextDecoder().decode(stdout);
    return output;
}

function getReleaseTypeFromCommits(): ReleaseType {
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
