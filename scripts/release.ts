import { inc, type ReleaseType } from "semver";
import { Octokit } from "@octokit/rest";
import pkgJson from "../package.json" with { type: "json" };
import { logRelease } from "./tool/constants.ts";
import { existsSync } from "node:fs";
import fs from "node:fs/promises";
import path, { basename } from "node:path";
import { spawnSync } from "node:child_process";
import isCI from "is-ci";
import { showHelp } from "./tool/help.ts";

if (
	showHelp({
		name: "release",
		description: "Calculate a release version and create a GitHub release.",
		usage: "bun run release [release-type] [alpha|beta] [--dry-run]",
		options: [
			"-h, --help                  Show this help.",
			"--dry-run                   Calculate the version and changelog without publishing.",
			"patch|minor|major|release   Select a release type; defaults to commit analysis.",
			"alpha|beta                  Create a prerelease; defaults to a prerelease major.",
		],
	})
) {
	process.exit(0);
}

const gitLogFormat = "- [%h] %s (@%cN)";
const preIdList = ["alpha", "beta"];
const args = process.argv.slice(2);
const preId = args.find((a) => preIdList.includes(a));

let isDryRun = args.includes("--dry-run");

if (!isCI) {
	isDryRun = true;
	logRelease.warn("This is not an actual release.");
	logRelease.warn(
		"You are not in CI environment, dry run is enabled by default to prevent accidental releases.\n" +
			"If you really want to release, run this script in CI environment on GitHub or set 'CI' environment variable to 'true' (not recommended).",
	);
}

if (!process.env.GITHUB_TOKEN) {
	logRelease.error("Environment variable GITHUB_TOKEN is not set.");
	process.exit(1);
}

const manifestPath = "package.json";
const manifestExists = existsSync(manifestPath);

if (!manifestExists) {
	logRelease.error("Manifest file " + manifestPath + " not found.");
	process.exit(1);
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
logRelease.info("Increment: " + pkgJson.version + " -> " + newVersion);
logRelease.info("Old latest tag: " + latestTagString);

const headRes = spawnSync("git", ["rev-parse", "HEAD"], { encoding: "utf-8" });
const preBumpCommitSha = headRes.stdout ? headRes.stdout.trim() : "";
let hasPushedBump = false;

try {
	// 1. If not a dry run and version changed, update, commit, and push package.json
	if (newVersion !== pkgJson.version) {
		const manifestOldContent = await fs.readFile(manifestPath, "utf-8");
		const manifestNewContent = manifestOldContent.replace(
			/("version"\s*:\s*")([^"]+)(")/,
			`$1${newVersion}$3`,
		);
		if (!isDryRun) {
			const task = logRelease.task({
				text: "Writing new version to " + manifestPath,
			});
			await fs.writeFile(manifestPath, manifestNewContent, "utf-8");
			task.end("completed");

			const taskGit = logRelease
				.task({
					text: "Committing and pushing version bump to git",
				})
				.start();
			commitAndPushManifest(newVersion);
			hasPushedBump = true;
			taskGit.end("completed");
		}
	}

	// 2. If not a dry run, compile client assets and cross-compile backend binaries
	if (!isDryRun) {
		const taskFront = logRelease
			.task({
				text: "Compiling client frontend assets",
			})
			.start();
		const frontCmd = spawnSync("bun", ["run", "front"], { stdio: "inherit" });
		if (frontCmd.status !== 0) {
			taskFront.end("failed");
			throw new Error("Failed to compile client frontend assets.");
		}
		taskFront.end("completed");

		const taskBack = logRelease
			.task({
				text: "Building cross-compiled binaries",
			})
			.start();
		const backCmd = spawnSync("bun", ["run", "back:cross"], {
			stdio: "inherit",
		});
		if (backCmd.status !== 0) {
			taskBack.end("failed");
			throw new Error("Failed to build cross-compiled binaries.");
		}
		taskBack.end("completed");

		// 2.1 Compute SHA-256 checksums and write to dist/checksums.txt
		const taskChecksums = logRelease
			.task({
				text: "Generating SHA-256 checksums",
			})
			.start();
		await writeChecksumsFile();
		taskChecksums.end("completed");
	}
} catch (error) {
	const errMsg = error instanceof Error ? error.message : String(error);
	logRelease.error(`Release process failed during build: ${errMsg}`);
	if (hasPushedBump) {
		rollback(preBumpCommitSha);
	}
	process.exit(1);
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
	const githubRef = process.env.GITHUB_REF_NAME;
	if (githubRef) {
		return githubRef;
	}
	const res = spawnSync("git", ["rev-parse", "--abbrev-ref", "HEAD"], {
		encoding: "utf-8",
	});
	if (res.status === 0 && res.stdout) {
		const branch = res.stdout.trim();
		if (branch && branch !== "HEAD") {
			return branch;
		}
	}
	return "main"; // Fallback
}

function commitAndPushManifest(version: string): void {
	const addCmd = spawnSync("git", ["add", "package.json"], {
		stdio: "inherit",
	});
	if (addCmd.status !== 0) {
		logRelease.error("Failed to git add package.json");
		process.exit(1);
	}
	const commitCmd = spawnSync(
		"git",
		["commit", "-m", `chore(release): bump version to ${version}`],
		{ stdio: "inherit" },
	);
	if (commitCmd.status !== 0) {
		logRelease.error("Failed to git commit package.json");
		process.exit(1);
	}
	const branchName = getCurrentBranch();
	const pushCmd = spawnSync("git", ["push", "origin", `HEAD:${branchName}`], {
		stdio: "inherit",
	});
	if (pushCmd.status !== 0) {
		logRelease.error("Failed to git push package.json");
		process.exit(1);
	}
}

function rollback(targetSha: string): void {
	logRelease.warn("Initiating automated rollback...");
	spawnSync("git", ["reset", "--hard", targetSha], { stdio: "inherit" });
	const branchName = getCurrentBranch();
	const pushCmd = spawnSync(
		"git",
		["push", "origin", `HEAD:${branchName}`, "--force"],
		{ stdio: "inherit" },
	);
	if (pushCmd.status !== 0) {
		logRelease.error("Failed to force-push rollback commit.");
	} else {
		logRelease.success(
			"Rollback completed. Remote branch restored to previous state.",
		);
	}
}

async function writeChecksumsFile(): Promise<void> {
	let checksumsContent = "";
	if (existsSync("dist")) {
		const files = await fs.readdir("dist");
		for (const file of files) {
			if (!file.startsWith("server-")) continue;
			const filePath = path.join("dist", file);
			const stat = await fs.stat(filePath);
			if (!stat.isFile()) continue;
			const hash = await getSha256(filePath);
			checksumsContent += `${hash}  ${file}\n`;
		}
	}
	if (checksumsContent) {
		await fs.writeFile("dist/checksums.txt", checksumsContent, "utf-8");
	}
}

async function getSha256(filePath: string): Promise<string> {
	const data = await fs.readFile(filePath);
	const hashBuffer = await crypto.subtle.digest("SHA-256", data);
	const hashArray = Array.from(new Uint8Array(hashBuffer));
	const hashHex = hashArray
		.map((b) => b.toString(16).padStart(2, "0"))
		.join("");
	return hashHex;
}

type RepoInfo = { owner: string; repoName: string };

async function createRelease(
	name: string,
	tagName: string,
	changelog: string,
	{ owner, repoName }: RepoInfo,
): Promise<void> {
	const o = new Octokit({ auth: process.env.GITHUB_TOKEN! });
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
	if (existsSync("dist")) {
		const files = await fs.readdir("dist");
		for (const file of files) {
			if (file.startsWith("server-")) {
				assetsToUpload.push(path.join("dist", file));
			}
		}
	}
	if (existsSync("dist/checksums.txt")) {
		assetsToUpload.push("dist/checksums.txt");
	}

	for (const assetPath of assetsToUpload) {
		const assetName = basename(assetPath);
		const content = await fs.readFile(assetPath);

		await o.repos.uploadReleaseAsset({
			release_id: release.data.id,
			owner,
			repo: repoName,
			name: assetName,
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

function getRepoInfo(): RepoInfo {
	const res = spawnSync("git", ["config", "--get", "remote.origin.url"], {
		encoding: "utf-8",
	});

	const output = res.stdout ? res.stdout.trim() : "";
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

	const found = (["major", "minor", "patch", "release"] as ReleaseType[]).find(
		(a) => args.includes(a),
	);
	let result: string | null;
	if (preId !== undefined) {
		releaseType = found ? (`pre${found}` as ReleaseType) : "premajor";
		result = inc(pkgJson.version, releaseType, preId);
	} else {
		releaseType = found || getReleaseTypeFromCommits();
		result = inc(pkgJson.version, releaseType);
	}

	if (!result) throw new Error("Invalid inc result");
	return result;
}

function createTag(name: string, changelog: string): void {
	const tagCmd = spawnSync("git", ["tag", "-a", name, "-m", changelog], {
		stdio: "inherit",
	});
	if (tagCmd.status !== 0) {
		logRelease.error("Failed to git tag release version.");
		process.exit(1);
	}

	const pushCmd = spawnSync("git", ["push", "origin", "--tags"], {
		stdio: "inherit",
	});
	if (pushCmd.status !== 0) {
		logRelease.error("Failed to push git tags.");
		process.exit(1);
	}
}

function getNewChangelog(from?: string): string {
	const res = spawnSync(
		"git",
		[
			"log",
			"--all",
			"--reverse",
			`--format=${gitLogFormat}`,
			`${from ? `${from}..` : ""}HEAD`,
		],
		{ encoding: "utf-8" },
	);

	const output = res.stdout ? res.stdout : "";
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
		const isMajor =
			/^[a-zA-Z\d]+(\([a-zA-Z\d]+\))?!/g.test(message) ||
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
	const res = spawnSync(
		"git",
		["describe", "--tags", "--match=*", "--no-abbrev", "HEAD"],
		{ encoding: "utf-8" },
	);

	if (res.status !== 0 || !res.stdout) {
		return undefined;
	}

	return res.stdout.trim();
}

function getCommitMessages(from?: string): string {
	const formatStr = "%s";
	const res = spawnSync(
		"git",
		[
			"log",
			"--reverse",
			`--format=${formatStr}`,
			`${from ? `${from}..` : ""}HEAD`,
		],
		{ encoding: "utf-8" },
	);

	if (res.status !== 0 || !res.stdout) {
		return "";
	}

	return res.stdout;
}
