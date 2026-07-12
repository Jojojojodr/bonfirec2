#!/usr/bin/env python3

import json
import tarfile
import zipfile
import platform
from pathlib import Path
from urllib.parse import urlparse
from urllib.request import Request, urlopen
from urllib.error import HTTPError, URLError

def get_asset_filename() -> str:
	system = platform.system().lower()
	machine = platform.machine().lower()
	if system == "linux":
		if machine == "x86_64":
			return "bfc2-client-linux-amd64.tar.gz"
		elif machine == "aarch64":
			return "bfc2-client-linux-arm64.tar.gz"
		else:
			raise RuntimeError(f"Unsupported Linux architecture: {machine}")
	elif system == "darwin":
		if machine == "x86_64":
			return "bfc2-client-darwin-amd64.tar.gz"
		elif machine == "aarch64":
			return "bfc2-client-darwin-arm64.tar.gz"
		else:
			raise RuntimeError(f"Unsupported Darwin architecture: {machine}")
	elif system == "windows":
		return "bfc2-client-windows-amd64.zip"
	else:
		raise RuntimeError(f"Unsupported platform: {system}")

def get_latest_release_tag(owner: str, repo: str) -> str:
	api_url = f"https://api.github.com/repos/{owner}/{repo}/releases/latest"
	request = Request(
		api_url,
		headers={
			"Accept": "application/vnd.github+json",
			"User-Agent": "bonfirec2-client-download-script",
		},
	)

	try:
		with urlopen(request, timeout=15) as response:
			data = json.loads(response.read().decode("utf-8"))
	except HTTPError as exc:
		raise RuntimeError(f"GitHub API returned HTTP {exc.code} for {api_url}") from exc
	except URLError as exc:
		raise RuntimeError(f"Failed to reach GitHub API: {exc.reason}") from exc

	tag = data.get("tag_name")
	if not tag:
		raise RuntimeError("GitHub API response did not include tag_name")
	return tag

def get_asset_url(owner: str, repo: str, asset_filename: str) -> str:
	tag = get_latest_release_tag(owner, repo)
	return f"https://github.com/{owner}/{repo}/releases/download/{tag}/{asset_filename}"

def download_asset(asset_url: str, output_path: str | None = None) -> Path:
	"""Download an asset URL and return the saved file path."""
	if output_path is None:
		filename = Path(urlparse(asset_url).path).name
		if not filename:
			raise RuntimeError("Could not determine filename from asset URL")
		destination = Path(filename)
	else:
		destination = Path(output_path)

	destination.parent.mkdir(parents=True, exist_ok=True)
	request = Request(asset_url, headers={"User-Agent": "bonfirec2-client-download-script"})

	try:
		with urlopen(request, timeout=60) as response, destination.open("wb") as out_file:
			while True:
				chunk = response.read(1024 * 128)
				if not chunk:
					break
				out_file.write(chunk)
	except HTTPError as exc:
		raise RuntimeError(f"Asset download failed with HTTP {exc.code}: {asset_url}") from exc
	except URLError as exc:
		raise RuntimeError(f"Failed to download asset: {exc.reason}") from exc

	return destination

def _archive_output_dir(archive_path: Path) -> Path:
	name = archive_path.name
	for suffix in (".tar.gz", ".tar.bz2", ".tar.xz", ".tgz", ".zip", ".tar"):
		if name.endswith(suffix):
			return archive_path.parent / name[: -len(suffix)]
	return archive_path.parent / archive_path.stem

def _is_within_directory(base_dir: Path, target_path: Path) -> bool:
	base_resolved = base_dir.resolve()
	target_resolved = target_path.resolve()
	return str(target_resolved).startswith(str(base_resolved) + "/") or target_resolved == base_resolved

def extract_archive(archive_path: Path, output_dir: Path | None = None) -> Path:
	"""Extract a .zip or .tar* archive and return extraction directory."""
	if output_dir is None:
		output_dir = _archive_output_dir(archive_path)

	output_dir.mkdir(parents=True, exist_ok=True)

	if zipfile.is_zipfile(archive_path):
		try:
			with zipfile.ZipFile(archive_path) as zf:
				for member in zf.infolist():
					target_path = output_dir / member.filename
					if not _is_within_directory(output_dir, target_path):
						raise RuntimeError(f"Unsafe path in zip archive: {member.filename}")
				zf.extractall(output_dir)
		except zipfile.BadZipFile as exc:
			raise RuntimeError(f"Invalid zip archive: {archive_path}") from exc
		return output_dir

	if tarfile.is_tarfile(archive_path):
		try:
			with tarfile.open(archive_path) as tf:
				for member in tf.getmembers():
					target_path = output_dir / member.name
					if not _is_within_directory(output_dir, target_path):
						raise RuntimeError(f"Unsafe path in tar archive: {member.name}")
				tf.extractall(output_dir)
		except tarfile.TarError as exc:
			raise RuntimeError(f"Invalid tar archive: {archive_path}") from exc
		return output_dir

	raise RuntimeError(f"Unsupported archive format: {archive_path.name}")

if __name__ == "__main__":
	asset_filename = get_asset_filename()
	asset_url = get_asset_url("Jojojojodr", "bonfirec2", asset_filename)
	print(f"Latest release asset URL: {asset_url}")

	downloaded_file = download_asset(asset_url, f"./{asset_filename}")
	print(f"Downloaded to: {downloaded_file.resolve()}")

	extracted_dir = extract_archive(downloaded_file)
	print(f"Extracted to: {extracted_dir.resolve()}")
