package rotatepassphrase

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CmdRotatePassphrase rotate the encryption passphrase used by all sessions
type CmdRotatePassphrase struct {
	passphrase    string
	newPassphrase string
	dir           string

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCmdRotatePassphrase creates a command to re-encrypt all sessions with a new passphrase
func NewCmdRotatePassphrase(f *cmdutil.Factory) *CmdRotatePassphrase {
	ccmd := &CmdRotatePassphrase{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "rotatePassphrase",
		Short: "Rotate the session encryption passphrase",
		Long: heredoc.Doc(`
			Change the passphrase used to encrypt sensitive session information (e.g. passwords and tokens)
			without having to manually update each session.

			Every session file in the session home folder (or a custom folder given via --dir) is checked
			for encrypted values. Each value is decrypted using the current passphrase and re-encrypted
			using the new passphrase, and the reference key file (.key) is updated to match the new
			passphrase.

			Sessions which can not be decrypted with the current passphrase are left untouched and reported,
			so a failed or partial rotation never destroys existing values.
		`),
		Example: heredoc.Doc(`
			### Example 1: Rotate the passphrase interactively

			$ c8y sessions rotatePassphrase

			You will be prompted for the current passphrase and then for the new passphrase (with confirmation)

			### Example 2: Rotate the passphrase non-interactively

			$ C8Y_PASSPHRASE="old" c8y sessions rotatePassphrase --newPassphrase "new"

			### Example 3: Preview which files would be changed without modifying anything

			$ c8y sessions rotatePassphrase --dry

			### Example 4: Rotate sessions stored in a custom folder (instead of the session home folder)

			$ c8y sessions rotatePassphrase --dir /backups/c8y-sessions
		`),
		RunE: ccmd.RunE,
	}

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.passphrase, "passphrase", "", "Current passphrase. Read from env C8Y_PASSPHRASE or prompted if missing")
	cmd.Flags().StringVar(&ccmd.newPassphrase, "newPassphrase", "", "New passphrase. Prompted (with confirmation) if missing")
	cmd.Flags().StringVar(&ccmd.dir, "dir", "", "Directory to scan for session files. Defaults to the session home folder")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdRotatePassphrase) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	cs := n.factory.IOStreams.ColorScheme()

	//
	// 1. Resolve the folder to scan and its reference key file
	//
	sessionDir := cfg.GetSessionHomeDir()
	keyFile := cfg.KeyFile()
	if n.dir != "" {
		sessionDir, err = homedir.Expand(n.dir)
		if err != nil {
			return err
		}
		if absDir, err := filepath.Abs(sessionDir); err == nil {
			sessionDir = absDir
		}
		if info, err := os.Stat(sessionDir); err != nil || !info.IsDir() {
			return fmt.Errorf("session directory does not exist or is not a directory. dir=%s", sessionDir)
		}
		keyFile = filepath.Join(sessionDir, config.KeyFileName)
	}

	//
	// 2. Read the reference key file and verify the current passphrase against it.
	// A custom directory (e.g. a backup or copy) does not have to contain a key file,
	// in which case the passphrase can not be pre-validated, however each session value
	// is still only rewritten if it can be decrypted with the current passphrase.
	//
	secretText := ""
	keyFileContents, err := os.ReadFile(keyFile)
	if err != nil {
		if n.dir == "" {
			return fmt.Errorf("could not read the key file (%s), so there is no passphrase to rotate. %w", keyFile, err)
		}
	} else {
		secretText = strings.TrimSpace(string(keyFileContents))
		if cfg.SecureData.IsEncryptedBytes(keyFileContents) != 1 {
			return fmt.Errorf("key file (%s) does not contain encrypted reference text", keyFile)
		}
	}

	currentPassphrase := n.passphrase
	if currentPassphrase == "" {
		currentPassphrase = cfg.Passphrase
	}
	prompter := prompt.NewPrompt(log)
	if currentPassphrase == "" {
		currentPassphrase, err = prompter.Password("Enter current passphrase", "")
		if err != nil {
			return err
		}
	}

	referenceText := ""
	if secretText != "" {
		referenceText, err = cfg.SecureData.DecryptString(secretText, currentPassphrase)
		if err != nil {
			return fmt.Errorf("current passphrase is incorrect (it does not match the key file %s)", keyFile)
		}
	} else {
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s No key file found in %s. The current passphrase can not be pre-validated, sessions which do not match it will be skipped\n", cs.WarningIcon(), sessionDir)
	}

	//
	// 2. Get the new passphrase
	//
	newPassphrase := n.newPassphrase
	if newPassphrase == "" {
		newPassphrase, err = prompter.PasswordWithConfirm("new encryption passphrase", "Creating a new encryption key for sessions")
		if err != nil {
			return err
		}
	}
	if newPassphrase == "" {
		return fmt.Errorf("new passphrase must not be empty")
	}

	//
	// 3. Re-encrypt every session file
	//
	sessionFiles, err := findSessionFiles(cfg, log, sessionDir)
	if err != nil {
		return err
	}

	totalFiles := 0
	totalValues := 0
	failedFiles := make([]string, 0)

	for _, file := range sessionFiles {
		changed, err := rotateSessionFile(cfg, file, currentPassphrase, newPassphrase, cfg.DryRun())
		if err != nil {
			log.Warnf("Skipping session. file=%s, reason=%s", file, err)
			failedFiles = append(failedFiles, file)
			continue
		}
		if changed > 0 {
			totalFiles++
			totalValues += changed
			if cfg.DryRun() {
				fmt.Fprintf(n.factory.IOStreams.ErrOut, "DRY: Would re-encrypt %d value(s) in %s\n", changed, file)
			} else {
				log.Infof("Re-encrypted session values. file=%s, count=%d", file, changed)
			}
		}
	}

	//
	// 4. Update the reference key file (if one was found) so future passphrase checks use the new passphrase
	//
	if referenceText != "" && !cfg.DryRun() {
		newKeyText, err := cfg.SecureData.EncryptString(referenceText, newPassphrase)
		if err != nil {
			return fmt.Errorf("failed to create new key file contents. %w", err)
		}
		if err := os.WriteFile(keyFile, []byte(newKeyText), 0600); err != nil {
			return fmt.Errorf("failed to update key file. %w", err)
		}
	}

	//
	// 5. Summary
	//
	switch {
	case cfg.DryRun():
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "DRY: Would rotate passphrase. sessions=%d, values=%d\n", totalFiles, totalValues)
	case totalFiles == 0 && len(failedFiles) > 0:
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s No sessions were rotated. Check that the current passphrase is correct\n", cs.WarningIcon())
	case totalFiles == 0:
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s No encrypted values found in %s. Nothing to do\n", cs.SuccessIcon(), sessionDir)
	default:
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Passphrase rotated. sessions=%d, values=%d\n", cs.SuccessIcon(), totalFiles, totalValues)
	}
	if len(failedFiles) > 0 {
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s The following sessions could not be decrypted with the current passphrase and were left unchanged:\n", cs.WarningIcon())
		for _, file := range failedFiles {
			fmt.Fprintf(n.factory.IOStreams.ErrOut, "  %s\n", file)
		}
	}
	if os.Getenv(config.EnvPassphrase) != "" && !cfg.DryRun() {
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s %s is set in your environment. Update it to the new passphrase (or unset it) before activating a session\n", cs.WarningIcon(), config.EnvPassphrase)
	}
	return nil
}

// findSessionFiles returns the list of session files in the given folder,
// applying the same skip rules as the interactive session selection
// (ignoring the activity log, git internals, extensions and dot/settings files)
func findSessionFiles(cfg *config.Config, log *logger.Logger, srcdir string) ([]string, error) {
	files := make([]string, 0)

	err := filepath.WalkDir(srcdir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := strings.ToLower(info.Name())
			if name == strings.ToLower(config.ActivityLogDirName) || name == ".git" || name == "extensions" || path == cfg.ExtensionsDataDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(info.Name(), config.SettingsGlobalName+".") || strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		switch strings.ToLower(filepath.Ext(info.Name())) {
		case ".json", ".yaml", ".yml", ".toml":
			log.Infof("Found session file: %s", path)
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// rotateSessionFile re-encrypts all encrypted values in a single session file.
// The file is only written back if every encrypted value could be decrypted with
// the current passphrase, so a session is never left half-rotated.
func rotateSessionFile(cfg *config.Config, file string, currentPassphrase string, newPassphrase string, dryRun bool) (int, error) {
	v := viper.New()
	v.SetConfigFile(file)
	if err := v.ReadInConfig(); err != nil {
		return 0, fmt.Errorf("could not parse file. %w", err)
	}

	changed := 0
	for _, key := range v.AllKeys() {
		value, ok := v.Get(key).(string)
		if !ok {
			continue
		}
		if cfg.SecureData.IsEncrypted(value) != 1 {
			continue
		}

		decryptedValue, err := cfg.SecureData.DecryptString(value, currentPassphrase)
		if err != nil {
			return 0, fmt.Errorf("could not decrypt property %q with the current passphrase", key)
		}
		encryptedValue, err := cfg.SecureData.EncryptString(decryptedValue, newPassphrase)
		if err != nil {
			return 0, fmt.Errorf("could not re-encrypt property %q. %w", key, err)
		}
		v.Set(key, encryptedValue)
		changed++
	}

	if changed > 0 && !dryRun {
		if err := v.WriteConfig(); err != nil {
			return 0, fmt.Errorf("could not write file. %w", err)
		}
	}
	return changed, nil
}
