
# detect which shell it is being sourced in
C8Y_SHELL="sh"
if [ -n "${BASH_VERSION:-}" ]; then
    case "$BASH_VERSION" in
        1*|2*|3*)
            # older bash versions don't support completions
            # and other syntax which
            C8Y_SHELL="sh"
            ;;
        *)
            C8Y_SHELL="bash"
            ;;
    esac
elif [ -n "$ZSH_VERSION" ]; then
    C8Y_SHELL="zsh"
fi

# dot source the real shell value
case "$C8Y_SHELL" in
    sh)
        eval "$(c8y cli profile --shell __sh)"
        ;;
    bash)
        eval "$(c8y cli profile --shell __bash)"
        ;;
    zsh)
        eval "$(c8y cli profile --shell __zsh)"
        ;;
    *)
        # fallback to something sensible
        eval "$(c8y cli profile --shell __bash)" 2>/dev/null || eval "$(c8y cli profile --shell __sh)"
        ;;
esac
