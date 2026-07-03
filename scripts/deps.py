import subprocess

go_deps = [
    "github.com/air-verse/air@latest",
    "github.com/a-h/templ/cmd/templ@v0.3.1020",
    "github.com/go-task/task/v3/cmd/task@latest",
]

def installing_dependencies():
    for dep in go_deps:
        print(f"Installing Go dependency: {dep}...")
    
        try:
            subprocess.run(
                ["go", "install", dep],
                check=True
            )
            print(f"Successfully installed {dep}.")
        except subprocess.CalledProcessError as e:
            print(f"Failed to install {dep}: {e}")
            raise

def installing_npm_dependencies():
    try:
        subprocess.run(
            ["npm", "install"],
            check=True
        )
        print("Successfully installed npm dependencies.")
    except subprocess.CalledProcessError as e:
        print(f"Failed to install npm dependencies: {e}")
        raise

if __name__ == "__main__":
    print("Installing dependencies...")
    installing_dependencies()
    installing_npm_dependencies()
    print("All dependencies installed successfully.")