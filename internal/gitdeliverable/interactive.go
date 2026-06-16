package gitdeliverable

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func EditPlan(in io.Reader, out io.Writer, plan Plan) (Plan, bool, error) {
	reader := bufio.NewReader(in)
	for {
		PrintPlan(out, plan)
		fmt.Fprintln(out, "Interactive commands:")
		fmt.Fprintln(out, "- enter: accept plan")
		fmt.Fprintln(out, "- p: print plan")
		fmt.Fprintln(out, "- e <n> <message>: edit commit message")
		fmt.Fprintln(out, "- m <file> <n>: move file to deliverable")
		fmt.Fprintln(out, "- n <message>: create new deliverable")
		fmt.Fprintln(out, "- q: quit")
		fmt.Fprint(out, "> ")
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return plan, false, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return plan, true, nil
		}
		if line == "q" || line == "quit" {
			return plan, false, nil
		}
		if line == "p" || line == "print" {
			continue
		}
		if err := applyInteractiveCommand(&plan, line); err != nil {
			fmt.Fprintf(out, "error: %s\n\n", err)
		}
	}
}

func Confirm(in io.Reader, out io.Writer, prompt string) (bool, error) {
	reader := bufio.NewReader(in)
	fmt.Fprintf(out, "%s [y/N]: ", prompt)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func applyInteractiveCommand(plan *Plan, line string) error {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
	}
	switch fields[0] {
	case "e", "edit":
		if len(fields) < 3 {
			return fmt.Errorf("usage: e <n> <message>")
		}
		number, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("invalid deliverable number %q", fields[1])
		}
		return plan.EditMessage(number, strings.Join(fields[2:], " "))
	case "m", "move":
		if len(fields) != 3 {
			return fmt.Errorf("usage: m <file> <n>")
		}
		number, err := strconv.Atoi(fields[2])
		if err != nil {
			return fmt.Errorf("invalid deliverable number %q", fields[2])
		}
		return plan.MoveFile(fields[1], number)
	case "n", "new":
		if len(fields) < 2 {
			return fmt.Errorf("usage: n <message>")
		}
		return plan.AddDeliverable(strings.Join(fields[1:], " "))
	default:
		return fmt.Errorf("unknown command %q", fields[0])
	}
}
