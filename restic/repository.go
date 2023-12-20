package restic

const (
	passwordEnv            string = "RESTIC_PASSWORD"
	resticProgressFPS      string = "RESTIC_PROGRESS_FPS"
	resticProgressFPSValue string = "2"
)

type ResticRepository interface {
	Init() ([]byte, error)
	Backup() error
	Snapshots() ([]byte, error)
}

func countStringLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}

	if len(s) > 0 && s[len(s)-1] != '\n' {
		count++
	}

	return count
}
