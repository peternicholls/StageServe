# 20i Stack GUI Manager

The GUI is experimental and not fully developed yet. Prefer the shell commands for the most reliable workflow.

## 🚀 Usage

From any project directory, simply run:

```bash
20i-gui
```

This currently gives you an interactive menu with these options:

### 📋 Menu Options:

1. **🚀 Start Stack (current directory)**
   - Uses the current directory as your project root
   - Auto-detects project name from folder name
   - Prompts for custom web port (defaults to 80)
   - Loads `.20i-local` file if present for project-specific settings

2. **🛑 Stop Stack**
   - Shows list of running 20i stacks
   - Choose specific project to stop or stop all
   - Clean shutdown of containers

3. **📊 View Status**
   - Overview of all running Docker containers
   - List of active 20i projects

4. **📋 View Logs**
   - Shows running 20i stacks
   - Follow real-time logs for selected project
   - Press Ctrl+C to stop following

## 🎯 Current Use Cases:

- **Basic project switching** without remembering commands
- **Lightweight inspection** of what is running
- **Trying the experimental menu flow** while the shell commands remain primary

## 🛠 Integration with Existing Workflow:

Your existing aliases still work perfectly:
- `dcu` - Start stack (command line)
- `dcd` - Stop stack (command line) 
- `20i` - View status (command line)
- `20i-gui` - Interactive menu (new!)

## 💡 Pro Tips:

- **Dialog Support**: Install `dialog` package for prettier menus:
  ```bash
  brew install dialog
  ```

- **Project Settings**: Create `.20i-local` in your project root:

   ```bash
   export HOST_PORT=8080
   export PHP_VERSION=8.4
   export MYSQL_DATABASE=myproject_db
   ```

- **One-off CLI override**: Start with a different PHP version without editing project config:

   ```bash
   20i-up --php-version 8.4
   20i-up version=8.4
   ```

- **From Anywhere**: The `20i-gui` command works from any project directory

Use it as a secondary option alongside the shell workflow while the GUI is still being developed.
