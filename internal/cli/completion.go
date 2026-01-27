package cli

import (
	"os"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Completion(args []string) {
	fs := flag.NewFlagSet("completion", flag.ExitOnError)
	shell := fs.String("shell", "bash", "Shell type (bash|zsh|fish)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed completion --shell <shell>\n\n")
		logger.SafeFprintf(r.Stderr, "Generate shell completion scripts.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed completion --shell bash >> ~/.bash_completion\n")
		logger.SafeFprintf(r.Stderr, "  workshed completion --shell zsh > ~/.zsh/completion/_workshed\n")
	}

	if err := fs.Parse(args); err != nil {
		r.ExitFunc(1)
		return
	}

	switch *shell {
	case "bash":
		r.generateBashCompletion()
	case "zsh":
		r.generateZshCompletion()
	case "fish":
		r.generateFishCompletion()
	default:
		logger.SafeFprintf(r.Stderr, "Unsupported shell: %s (supported: bash, zsh, fish)\n", *shell)
		r.ExitFunc(1)
	}
}

func (r *Runner) generateBashCompletion() {
	logger.SafeFprintf(r.Stdout, `# workshed bash completion
_workshed_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="create list inspect path exec repos captures capture apply export remove update completion help version health"
    opts="--help --format --shell --name --repo --repos --yes"

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi

    if [[ ${COMP_CWORD} -eq 1 ]] ; then
        COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
        return 0
    fi

    if [[ "${prev}" == "exec" || "${prev}" == "path" || "${prev}" == "inspect" || "${prev}" == "remove" || "${prev}" == "apply" || "${prev}" == "repos" ]] ; then
        COMPREPLY=( $(compgen -W "$(workshed list --format raw 2>/dev/null)" -- ${cur}) )
        return 0
    fi

    return 0
}

complete -F _workshed_completions workshed
`)
}

func (r *Runner) generateZshCompletion() {
	logger.SafeFprintf(r.Stdout, `#compdef workshed

local -a commands opts
commands=(
    'create:Create a new workspace'
    'list:List workspaces'
    'inspect:Show workspace details'
    'path:Show workspace path'
    'exec:Run a command in repositories'
    'repos:Manage repositories'
    'captures:List captures'
    'capture:Create a capture'
    'apply:Apply a captured state'
    'export:Export workspace configuration'
    'remove:Remove a workspace'
    'update:Update workspace purpose'
    'completion:Generate shell completion'
    'health:Check workspace health'
    'help:Show help'
    'version:Show version'
)

opts=(
    '--help[Show help]'
    '--format[Output format]:format:(table json raw)'
    '--shell[Shell type]:shell:(bash zsh fish)'
    '--name[Capture name]'
    '--repo[Repository URL]'
    '--repos[Repository URLs]'
    '--yes[Skip confirmation]'
)

_workshed() {
    local -a workspace_handles
    if [[ ${#words[@]} -ge 2 ]]; then
        workspace_handles=( ${(f)"$(workshed list --format raw 2>/dev/null)"} )
    fi

    if [[ ${CURRENT} -eq 2 ]] ; then
        _describe 'command' commands
        return
    fi

    case ${words[2]} in
        exec|path|inspect|remove|apply|repos|captures|capture|export)
            _describe 'workspace' workspace_handles
            ;;
    esac
}

_workshed
`)
}

func (r *Runner) generateFishCompletion() {
	logger.SafeFprintf(r.Stdout, `# workshed fish completion
complete -c workshed -f -a "(workshed list --format raw 2>/dev/null)"

complete -c workshed -n "__fish_use_subcommand" -a create -d "Create a new workspace"
complete -c workshed -n "__fish_use_subcommand" -a list -d "List workspaces"
complete -c workshed -n "__fish_use_subcommand" -a inspect -d "Show workspace details"
complete -c workshed -n "__fish_use_subcommand" -a path -d "Show workspace path"
complete -c workshed -n "__fish_use_subcommand" -a exec -d "Run a command in repositories"
complete -c workshed -n "__fish_use_subcommand" -a repos -d "Manage repositories"
complete -c workshed -n "__fish_use_subcommand" -a captures -d "List captures"
complete -c workshed -n "__fish_use_subcommand" -a capture -d "Create a capture"
complete -c workshed -n "__fish_use_subcommand" -a apply -d "Apply a captured state"
complete -c workshed -n "__fish_use_subcommand" -a export -d "Export workspace configuration"
complete -c workshed -n "__fish_use_subcommand" -a remove -d "Remove a workspace"
complete -c workshed -n "__fish_use_subcommand" -a update -d "Update workspace purpose"
complete -c workshed -n "__fish_use_subcommand" -a completion -d "Generate shell completion"
complete -c workshed -n "__fish_use_subcommand" -a health -d "Check workspace health"
complete -c workshed -n "__fish_use_subcommand" -a help -d "Show help"
complete -c workshed -n "__fish_use_subcommand" -a version -d "Show version"

complete -c workshed -l help -d "Show help"
complete -c workshed -l format -d "Output format" -r -a "table json raw"
complete -c workshed -l shell -d "Shell type" -r -a "bash zsh fish"
complete -c workshed -l name -d "Capture name"
complete -c workshed -l repo -d "Repository URL"
complete -c workshed -l repos -d "Repository URLs"
complete -c workshed -s y -l yes -d "Skip confirmation"
`)
}

func init() {
	_ = os.Setenv("WORKSHED_LOG_FORMAT", "raw")
}
