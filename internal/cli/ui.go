package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"moon-prism-power/internal/migration"
)

func planSummary(plan migration.Plan) string {
	var create, update, skip int
	var details strings.Builder
	for _, job := range plan.Jobs {
		switch job.Action {
		case migration.ActionCreate:
			create++
		case migration.ActionUpdate:
			update++
		case migration.ActionSkip:
			skip++
			continue
		}

		fmt.Fprintf(&details, "  - %s %s %q (MAL ID %d)", strings.ToUpper(string(job.Action)), job.Entry.Kind, job.Entry.Title, job.Entry.MALID)
		if job.Current != nil {
			fmt.Fprintf(&details, ": before {%s} -> after {%s}", targetSummary(*job.Current), targetSummary(job.Update))
		} else {
			fmt.Fprintf(&details, ": after {%s}", targetSummary(job.Update))
		}

		details.WriteByte('\n')
	}

	planned := details.String()
	if planned == "" {
		planned = "  No changes required.\n"
	}

	return fmt.Sprintf("Preview: %d create, %d update, %d skip.\nPlanned changes:\n%s", create, update, skip, planned)
}

func targetSummary(update migration.TargetUpdate) string {
	return fmt.Sprintf("status=%s, score=%d, progress=%d, volumes=%d, repeat=%d, dates=%s..%s", update.Status, update.Score, update.Progress, update.Volumes, update.Repeat, update.StartDate, update.FinishDate)
}

func confirm(input io.Reader, output io.Writer) bool {
	_, _ = fmt.Fprint(output, "This will update your MyAnimeList list. Continue? [y/N] ")
	answer, _ := bufio.NewReader(input).ReadString('\n')
	answer = strings.TrimSpace(answer)
	return strings.EqualFold(answer, "y") || strings.EqualFold(answer, "yes")
}
