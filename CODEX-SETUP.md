# CODEX Environment Setup

**Quick setup for WebSocket Service bug fixing in CODEX environment**

## 🚀 One-Time Setup Instructions

### 1. Configure CODEX Preinstalled Packages
When creating a new CODEX environment, select:
- **Go**: `1.23.8` (or latest available)
- **Python**: `3.12` (for any tooling)
- **Node.js**: `20` (for any web tooling)

### 2. Run Setup Script
```bash
# Navigate to project directory
cd /workspace

# Run the setup script
./codex-setup.sh
```

## ⚡ What the Script Does
1. ✅ Sets up Go environment (GOPATH, GOBIN)
2. ✅ Installs all required Go tools (protoc plugins, staticcheck, air, etc.)
3. ✅ Downloads project dependencies
4. ✅ Generates protobuf code (fixes Bug #01)
5. ✅ Tests build and runs CI verification
6. ✅ Ready for bug fixing!

## 🐛 Start Bug Fixing
After setup completes, follow the workflow in `AGENTS.md`:

```bash
# Pick a bug and create branch
git checkout -b bug/03-race-condition-websocket-handler

# Follow the bug report fix instructions
# Test your changes
make ci

# Commit and push
git add .
git commit -s -m "fix(websocket): race in broadcast map (03-bug-03)"
git push origin bug/03-race-condition-websocket-handler
```

## 🔧 Development Commands
- `make dev` - Hot reload development
- `make ci` - Full CI pipeline  
- `make test-race` - Race condition testing
- `make build` - Build the service
- `make proto` - Generate protobuf code

## 📚 Documentation
- `AGENTS.md` - Complete bug-fixing workflow
- `docs/development/branch-conventions.md` - Branch naming rules
- `.github/PULL_REQUEST_TEMPLATE.md` - PR checklist

---

**Setup time: ~1 minute**  
**Ready to fix 12 documented bugs!** 🎯 