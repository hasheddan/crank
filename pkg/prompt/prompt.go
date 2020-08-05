/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package prompt

import (
	"bufio"
	"fmt"
)

// Output colors.
const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
)

// FmtInfo formats an info message.
func FmtInfo(s string) string {
	return fmt.Sprintf(InfoColor, s)
}

// FmtNotice formats an notice message.
func FmtNotice(s string) string {
	return fmt.Sprintf(NoticeColor, s)
}

// FmtWarning formats a warning message.
func FmtWarning(s string) string {
	return fmt.Sprintf(WarningColor, s)
}

// FmtError formats an error message.
func FmtError(s string) string {
	return fmt.Sprintf(ErrorColor, s)
}

// FmtDebug formats a debug message.
func FmtDebug(s string) string {
	return fmt.Sprintf(DebugColor, s)
}

// A Step is a step in prompter execution.
type Step int

// A Fn prompts a user for input.
type Fn func(p Prompter, s string) (string, bool)

// BareFn is a Fn that requires no input.
func BareFn(p func() (string, bool)) Fn {
	return func(_ Prompter, _ string) (string, bool) {
		return p()
	}
}

// BinaryFn is a Fn that expects a yes or no answer.
func BinaryFn(f func(bool, Prompter, string) (string, bool)) Fn {
	return func(p Prompter, s string) (string, bool) {
		if s == "Y" {
			return f(true, p, s)
		}
		return f(false, p, s)
	}
}

// Prompter is the interface that must be satisfied by a Prompter.
type Prompter interface {
	Prompt() error
	SetStep(Step)
}

// prompter is an interactive command prompt.
type prompter struct {
	initialPrompt string
	scanner       *bufio.Scanner
	steps         map[Step]Fn
	step          Step
}

// NewPrompter returns a new prompter.
func NewPrompter(initial string, scanner *bufio.Scanner, steps map[Step]Fn, firstStep Step) Prompter {
	return &prompter{
		initialPrompt: initial,
		scanner:       scanner,
		steps:         steps,
		step:          firstStep,
	}
}

// SetStep sets the current prompter step.
func (p *prompter) SetStep(s Step) {
	p.step = s
}

// Prompt runs a Prompter.
func (p *prompter) Prompt() error {
	fmt.Printf("%s\n--> ", p.initialPrompt)
	for {
		p.scanner.Scan()
		t := p.scanner.Text()
		out, done := p.steps[p.step](p, t)
		if done {
			fmt.Println(out)
			break
		}
		fmt.Printf("%s\n--> ", out)
	}
	return nil
}
