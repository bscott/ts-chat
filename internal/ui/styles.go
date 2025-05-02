package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color definitions
var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#2B5F3A"}
	accent    = lipgloss.AdaptiveColor{Light: "#1D9BF0", Dark: "#1D9BF0"}
	warning   = lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#F25D94"}
)

// Style definitions
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(subtle)

	// Headers
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(highlight).
		Padding(0, 1)

	// Text styles
	SystemStyle = lipgloss.NewStyle().
		Foreground(special).
		Bold(true)

	UserStyle = lipgloss.NewStyle().
		Foreground(accent).
		Bold(true)

	SelfStyle = lipgloss.NewStyle().
		Foreground(highlight).
		Bold(true)

	ActionStyle = lipgloss.NewStyle().
		Foreground(warning).
		Italic(true)

	// UI components
	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(subtle).
		Padding(0, 1)

	InputStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(highlight).
		Padding(0, 1)
)

// FormatSystemMessage formats a system message
func FormatSystemMessage(message string) string {
	return SystemStyle.Render("[System] " + message)
}

// FormatUserMessage formats a user message
func FormatUserMessage(username, message, timestamp string) string {
	return UserStyle.Render("["+timestamp+"] "+username+": ") + message
}

// FormatSelfMessage formats the user's own message
func FormatSelfMessage(message, timestamp string) string {
	return SelfStyle.Render("["+timestamp+"] You: ") + message
}

// FormatActionMessage formats an action message
func FormatActionMessage(username, action string) string {
	return ActionStyle.Render("* " + username + " " + action)
}

// FormatTitle formats a title
func FormatTitle(title string) string {
	return HeaderStyle.Render("=== " + title + " ===")
}

// CreateColoredBox creates a colored box with a title and content
func CreateColoredBox(title, content string, width int) string {
	box := BoxStyle.Copy().Width(width)
	return box.Render(
		HeaderStyle.Render(title) + "\n\n" +
		content,
	)
}

// FormatHelp formats the help message
func FormatHelp() string {
	return BoxStyle.Render(
		HeaderStyle.Render("Available Commands:") + "\n" +
			"/who - Show all users in the room\n" +
			"/me <action> - Perform an action\n" +
			"/help - Show this help message\n" +
			"/quit - Leave the chat",
	)
}

// FormatUserList formats the user list
func FormatUserList(roomName string, users []string, maxUsers int) string {
	content := HeaderStyle.Render("Users in "+roomName+" ("+lipgloss.NewStyle().Foreground(accent).Render(fmt.Sprintf("%d/%d", len(users), maxUsers))+"):") + "\n"
	
	for _, user := range users {
		content += "- " + UserStyle.Render(user) + "\n"
	}
	
	return BoxStyle.Render(content)
}

// FormatWelcomeMessage formats the welcome message
func FormatWelcomeMessage(roomName, nickname string) string {
	return HeaderStyle.Render("Welcome to "+roomName+", "+nickname+"!") + "\n\n" +
		"Type a message and press Enter to send. Use /help to see available commands."
}