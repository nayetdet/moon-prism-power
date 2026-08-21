package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"moon-prism-power/internal/migration"
)

func summary(plan migration.Plan) string {
	var create, update, skip int
	for _, job := range plan.Jobs {
		switch job.Action {
		case migration.ActionCreate:
			create++
		case migration.ActionUpdate:
			update++
		case migration.ActionSkip:
			skip++
		}
	}

	return fmt.Sprintf("Preview: %d create, %d update, %d skip.\n", create, update, skip)
}

func confirm(input io.Reader, output io.Writer) bool {
	fmt.Fprint(output, "This will update your MyAnimeList list. Continue? [y/N] ")
	answer, _ := bufio.NewReader(input).ReadString('\n')
	answer = strings.TrimSpace(answer)
	return strings.EqualFold(answer, "y") || strings.EqualFold(answer, "yes")
}
