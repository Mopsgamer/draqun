import { inc, type ReleaseType } from "semver";
import { Octokit } from "@octokit/rest";
import denojson from "../deno.json" with { type: "json" };
import { logRelease, taskDotenv } from "./tool/constants.ts";
import { existsSync, expandGlob } from "@std/fs";
import { basename } from "@std/path";
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
	logRelease.error("Environment variable GITHUB_TOKEN is not set.");
	Deno.exit(1);
}

const manifestPath = "deno.json";
const manifestExists = existsSync(manifestPath);

if (!manifestExists) {
	logRelease.error("Manifest file " + manifestPath + " not found.");
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

const { stdout: headStdout } = new Deno.Command("git", {
	args: ["rev-parse", "HEAD"],
}).outputSync();
const preBumpCommitSha = new TextDecoder().decode(headStdout).trim();
let hasPushedBump = false;

try {
	// 1. If not a dry run and version changed, update, commit, and push deno.json
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

			const taskGit = logRelease.task({
				text: "Committing and pushing version bump to git",
			}).start();
			commitAndPushManifest(newVersion);
			hasPushedBump = true;
			taskGit.end("completed");
		}
	}

	// 2. If not a dry run, compile client assets and cross-compile backend binaries
	if (!isDryRun) {
		const taskFront = logRelease.task({
			text: "Compiling client frontend assets",
		}).start();
		const frontCmd = new Deno.Command("deno", {
			args: ["task", "front"],
		}).outputSync();
		if (!frontCmd.success) {
			taskFront.end("failed");
			throw new Error("Failed to compile client frontend assets.");
		}
		taskFront.end("completed");

		const taskBack = logRelease.task({
			text: "Building cross-compiled binaries",
		}).start();
		const backCmd = new Deno.Command("deno", {
			args: ["task", "back:cross"],
		}).outputSync();
		if (!backCmd.success) {
			taskBack.end("failed");
			throw new Error("Failed to build cross-compiled binaries.");
		}
		taskBack.end("completed");

		// 2.1 Compute SHA-256 checksums and write to dist/checksums.txt
		const taskChecksums = logRelease.task({
			text: "Generating SHA-256 checksums",
		}).start();
		await writeChecksumsFile();
		taskChecksums.end("completed");
	}
} catch (error) {
	const errMsg = error instanceof Error ? error.message : String(error);
	logRelease.error(`Release process failed during build: ${errMsg}`);
	if (hasPushedBump) {
		rollback(preBumpCommitSha);
	}
	Deno.exit(1);
}

// 3. If not a dry run, create tag and push it
if (!isDryRun) {
	const task = logRelease.task({ text: "Creating tag " + tagName }).start();
	createTag(tagName, newChangelog);
	task.end("completed");
}

// 4. If not a dry run, create GitHub release and upload binaries and checksums
if (!isDryRun) {
	const repoInfo = getRepoInfo();
	const task = logRelease.task({ text: "Creating GitHub release" }).start();
	await createRelease(releaseName, tagName, newChangelog, repoInfo);
	task.end("completed");
	logRelease.success("Released successfully.");
}

function getCurrentBranch(): string {
	const githubRef = Deno.env.get("GITHUB_REF_NAME");
	if (githubRef) {
		return githubRef;
	}
	const { success, stdout } = new Deno.Command("git", {
		args: ["rev-parse", "--abbrev-ref", "HEAD"],
	}).outputSync();
	if (success) {
		const branch = new TextDecoder().decode(stdout).trim();
		if (branch && branch !== "HEAD") {
			return branch;
		}
	}
	return "main"; // Fallback
}

function commitAndPushManifest(version: string): void {
	const addCmd = new Deno.Command("git", {
		args: ["add", "deno.json"],
	}).outputSync();
	if (!addCmd.success) {
		logRelease.error("Failed to git add deno.json");
		Deno.exit(1);
	}
	const commitCmd = new Deno.Command("git", {
		args: ["commit", "-m", `chore(release): bump version to ${version}`],
	}).outputSync();
	if (!commitCmd.success) {
		logRelease.error("Failed to git commit deno.json");
		Deno.exit(1);
	}
	const branchName = getCurrentBranch();
	const pushCmd = new Deno.Command("git", {
		args: ["push", "origin", `HEAD:${branchName}`],
	}).outputSync();
	if (!pushCmd.success) {
		logRelease.error("Failed to git push deno.json");
		Deno.exit(1);
	}
}

function rollback(targetSha: string): void {
	logRelease.warn("Initiating automated rollback...");
	new Deno.Command("git", {
		args: ["reset", "--hard", targetSha],
	}).outputSync();
	const branchName = getCurrentBranch();
	const pushCmd = new Deno.Command("git", {
		args: ["push", "origin", `HEAD:${branchName}`, "--force"],
	}).outputSync();
	if (!pushCmd.success) {
		logRelease.error("Failed to force-push rollback commit.");
	} else {
		logRelease.success(
			"Rollback completed. Remote branch restored to previous state.",
		);
	}
}

async function writeChecksumsFile(): Promise<void> {
	let checksumsContent = "";
	for await (const entry of expandGlob("dist/server-*")) {
		if (!entry.isFile) continue;
		const hash = await getSha256(entry.path);
		checksumsContent += `${hash}  ${entry.name}\n`;
	}
	if (checksumsContent) {
		await Deno.writeTextFile("dist/checksums.txt", checksumsContent);
	}
}

async function getSha256(filePath: string): Promise<string> {
	const data = await Deno.readFile(filePath);
	const hashBuffer = await crypto.subtle.digest("SHA-256", data);
	const hashArray = Array.from(new Uint8Array(hashBuffer));
	const hashHex = hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
	return hashHex;
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

	const assetsToUpload: string[] = [];
	for await (const entry of expandGlob("dist/server-*")) {
		if (entry.isFile) assetsToUpload.push(entry.path);
	}
	if (existsSync("dist/checksums.txt")) {
		assetsToUpload.push("dist/checksums.txt");
	}

	for (const assetPath of assetsToUpload) {
		const name = basename(assetPath);
		const content = await Deno.readFile(assetPath);

		await o.repos.uploadReleaseAsset({
			release_id: release.data.id,
			owner,
			repo: repoName,
			name,
			data: content as unknown as string,
			headers: {
				"content-type": "application/octet-stream",
				"content-length": content.byteLength,
			},
		});
	}

	await o.repos.updateRelease({
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
	const tagCmd = new Deno.Command(`git`, {
		args: [
			"tag",
			"-a",
			name,
			`-m`,
			changelog,
		],
	}).outputSync();
	if (!tagCmd.success) {
		logRelease.error("Failed to git tag release version.");
		Deno.exit(1);
	}

	const pushCmd = new Deno.Command(`git`, {
		args: [
			"push",
			"origin",
			`--tags`,
		],
	}).outputSync();
	if (!pushCmd.success) {
		logRelease.error("Failed to push git tags.");
		Deno.exit(1);
	}
}

function getNewChangelog(from?: string): string {
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
	const lines = output.split("\n").filter((l) => l.trim() !== "");
	if (lines.length === 0) {
		return "Initial release.";
	}

	const categories: Record<string, string[]> = {
		"Features 🚀": [],
		"Bug Fixes 🐛": [],
		"Documentation 📝": [],
		"Performance Improvements ⚡": [],
		"Refactors & Improvements ⚙️": [],
		"Other Changes 🧹": [],
	};

	for (const line of lines) {
		// Example: "- [ab12cd3] feat(scope): add new login screen (@user)"
		const match = line.match(
			/^-\s*\[[0-9a-fA-F]+\]\s+([a-zA-Z]+)(?:\([a-zA-Z0-9._-]+\))?!?:\s*(.*)/,
		);
		if (match) {
			const type = match[1].toLowerCase();
			if (type === "feat") {
				categories["Features 🚀"].push(line);
			} else if (type === "fix") {
				categories["Bug Fixes 🐛"].push(line);
			} else if (type === "docs") {
				categories["Documentation 📝"].push(line);
			} else if (type === "perf") {
				categories["Performance Improvements ⚡"].push(line);
			} else if (type === "refactor" || type === "style") {
				categories["Refactors & Improvements ⚙️"].push(line);
			} else {
				categories["Other Changes 🧹"].push(line);
			}
		} else {
			categories["Other Changes 🧹"].push(line);
		}
	}

	let formattedChangelog = "";
	for (const [title, commitLines] of Object.entries(categories)) {
		if (commitLines.length > 0) {
			formattedChangelog += `### ${title}\n\n`;
			for (const commitLine of commitLines) {
				formattedChangelog += `${commitLine}\n`;
			}
			formattedChangelog += "\n";
		}
	}

	return formattedChangelog.trim() || output;
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
