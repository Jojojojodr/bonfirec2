import config
import deps

def main():
    print("Copying config file...")
    config.copy_config_file(config.src, config.dest)

    print("Installing Go dependencies...")
    deps.installing_dependencies()

    print("Installing npm dependencies...")
    deps.installing_npm_dependencies()

if __name__ == "__main__":
    print("Setting up the project...")
    main()
    print("Setup completed successfully.")