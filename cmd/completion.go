package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

const (

	/*
		__custom_func is called automatically when none of the auto-created (via cobra)
		functions handle the given input

		Currently the only cases handled are the following

		'odo create':
		Handled by simply providing a list of name of the available components
		(meaning no namespaces or versions are shown).
		- 'awk' is first used in order to filter out the first line containing the "header" and any
		trailing lines that might exist after the data.
		- Then 'awk' is used again to use select only the name.
		- Finally 'paste' is used turn the multiple lines into a single line of names separated by spaces

		'odo service create':
		Handled by providing the available services
		- 'cut' is used in order to remove the leading characters from the service names
		- 'paste' is then used turn the multiple lines into a single line of names separated by spaces

		'odo project delete|set':
		Handled by providing the available projects
		- 'tail' is used in order to drop the first line which is sort of a "header"
		- 'sed' is used to ensure that we only keep the project name
		- 'paste' is then used turn the multiple lines into a single line of names separated by spaces

		'odo storage delete|mount|unmount':
		Handled by providing the available storage objects
		It should be noted that this only works for the current component
		- 'awk' is first used in order to filter out the first line containing the "header" and any
		trailing lines that might exist after the first table.
		- 'sed' is used to filter out any informative messages that might exist
		- 'sed' is used again to remove empty lines
		- Then 'awk' is used again to use select only the name.
		- Finally 'paste' is used turn the multiple lines into a single line of names separated by spaces

		More information about writing bash completion functions can be found at
		https://debian-administration.org/article/317/An_introduction_to_bash_completion_part_2 for
	*/

	bashCompletionFunc = `
__custom_func() {
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"

    if [ "${#COMP_WORDS[@]}" -eq "3" ]; then # no entity has been entered, only a command like 'odo create'
      local command="${COMP_WORDS[COMP_CWORD-1]}"
      case "${command}" in
    		create)
    			local components=$(odo catalog list components | awk  '/NAME/{flag=1;next}/---/{flag=0}flag' | awk '{ print $1; }' | paste -sd " " -)
    			COMPREPLY=( $(compgen -W "${components}" -- ${cur}) )
          return 0
          ;;
      esac
    elif [ "${#COMP_WORDS[@]}" -eq "4" ]; then # an entity followed by a command has been entered like 'odo service create'
      local entity="${COMP_WORDS[COMP_CWORD-2]}"
      local verb="${COMP_WORDS[COMP_CWORD-1]}"
      case "${entity}" in
        service)
          case "${verb}" in
            create)
              local services=$(odo catalog list services | cut -c 3- | paste -sd " " -)
              COMPREPLY=( $(compgen -W "${services}" -- ${cur}) )
              return 0
              ;;
          esac
          ;;
        project)
          case "${verb}" in
            delete|set)
              local projects=$(odo project list | tail -n +2 | sed 's/[^-a-zA-Z0-9]//g' | paste -sd " " -)
              COMPREPLY=( $(compgen -W "${projects}" -- ${cur}) )
              return 0
              ;;
          esac
          ;;
        storage)
          case "${verb}" in
            delete|mount|unmount)
              local storages=$(odo storage list | awk  '/NAME/{flag=1;next}/NAME/{flag=0}flag' | sed '/No/d' | sed '/^\s*$/d' | awk '{ print $1; }' | paste -sd " " -)
              COMPREPLY=( $(compgen -W "${storages}" -- ${cur}) )
              return 0
              ;;
          esac
          ;;
      esac
    fi

  	return 0;
}
`
)

var completionCmd = &cobra.Command{
	Use:   "completion SHELL",
	Short: "Output shell completion code",
	Long: `Generates shell completion code.

Auto completion supports both bash and zsh. Output is to STDOUT.`,

	Example: `  # Bash autocompletion support
  source <(odo utils completion bash)

  # Zsh autocompletion support
  source <(odo utils completion zsh)
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		err := Generate(cmd, args)
		checkError(err, "")

		return nil
	},
}

// Generate the appropriate autocompletion file
func Generate(cmd *cobra.Command, args []string) error {

	// Check the passed in arguments
	if len(args) == 0 {
		return fmt.Errorf("Shell not specified. ex. odo completion [bash|zsh]")
	}
	if len(args) > 1 {
		return fmt.Errorf("Too many arguments. Expected only the shell type. ex. odo completion [bash|zsh]")
	}
	shell := args[0]

	// Generate bash through cobra if selected
	if shell == "bash" {
		return cmd.Root().GenBashCompletion(os.Stdout)

		// Generate zsh with the appropriate conversion as well as bash inclusion
	} else if shell == "zsh" {
		return runCompletionZsh(os.Stdout, cmd.Root())

		// Else, return an error.
	} else {
		return fmt.Errorf("not a compatible shell, bash and zsh are only supported")
	}
}

/*
	This is copied from
	https://github.com/kubernetes/kubernetes/blob/ea18d5c32ee7c320fe96dda6b0c757476908e696/pkg/kubectl/cmd/completion.go
	in order to generate ZSH completion support.
*/
func runCompletionZsh(out io.Writer, odo *cobra.Command) error {

	zshInitialization := `
__odo_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__odo_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__odo_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__odo_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__odo_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__odo_declare() {
	if [ "$1" == "-F" ]; then
		whence -w "$@"
	else
		builtin declare "$@"
	fi
}
__odo_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__odo_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__odo_filedir() {
	local RET OLD_IFS w qw
	__debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__odo_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__odo_quote() {
    if [[ $1 == \'* || $1 == \"* ]]; then
        # Leave out first character
        printf %q "${1:1}"
    else
    	printf %q "$1"
    fi
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__odo_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__odo_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__odo_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__odo_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__odo_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__odo_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/__odo_declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__odo_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	out.Write([]byte(zshInitialization))

	buf := new(bytes.Buffer)
	odo.GenBashCompletion(buf)
	out.Write(buf.Bytes())

	zshTail := `
BASH_COMPLETION_EOF
}
__odo_bash_source <(__odo_convert_bash_to_zsh)
`
	out.Write([]byte(zshTail))
	return nil
}
