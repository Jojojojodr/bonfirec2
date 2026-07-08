import shutil
from pathlib import Path

src = Path("config.example.yaml")
dest = Path("config.yaml")

def copy_config_file(src: Path, dest: Path) -> None:
    if dest.exists():
        return

    dest.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dest)

if __name__ == "__main__":
    print(f"Copying config file from {src} to {dest}...")
    copy_config_file(src, dest)
    print("Config file copied successfully.")
