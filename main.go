package runcheck

import (
	"fmt"
	"github.com/spf13/afero"
	"os"
	"time"
)

func main() {
	fs := afero.NewOsFs()
	SetFs(fs)
	abort, err := AbortStartup(".", "runConfig")

	if err != nil {
		fmt.Println(err.Error())
	}

	if abort {
		fmt.Println("Already running out of this directory. Exiting.")
		os.Exit(0)
	}

	fmt.Print("Waiting ...")
	time.Sleep(30 * time.Second)
}
