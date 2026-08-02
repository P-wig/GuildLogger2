package discord

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Interaction type constants (Discord API).
const (
	InteractionTypePing                           = 1
	InteractionTypeApplicationCommand             = 2
	InteractionTypeMessageComponent               = 3
	InteractionTypeApplicationCommandAutocomplete = 4
	InteractionTypeModalSubmit                    = 5
)

// Interaction response type constants (Discord API).
const (
	InteractionResponsePong                   = 1
	InteractionResponseChannelMessage         = 4
	InteractionResponseDeferredChannelMessage = 5 // ACK with ephemeral loading state; follow up within 15 min
	InteractionResponseDeferredUpdate         = 6
	InteractionResponseUpdateMessage          = 7
	InteractionResponseAutocomplete           = 8
	InteractionResponseModal                  = 9
)

// MessageFlagEphemeral makes the response only visible to the invoking user.
const MessageFlagEphemeral = 64

// --- Component + embed types ---

// Embed is a Discord rich embed object attached to a channel message.
type Embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Color       int          `json:"color,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
}

// EmbedField is a named field inside an Embed.
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// EmbedFooter is a footer text line inside an Embed.
type EmbedFooter struct {
	Text string `json:"text"`
}

// ActionRow is a container component (type 1) that holds buttons or selects.
type ActionRow struct {
	Type       int         `json:"type"` // always 1
	Components []Component `json:"components"`
}

// Component is a button (type 2) inside an ActionRow.
type Component struct {
	Type     int    `json:"type"`
	Style    int    `json:"style,omitempty"` // 1 primary, 2 secondary, 3 success, 4 danger
	Label    string `json:"label,omitempty"`
	CustomID string `json:"custom_id,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

// TextInputRow is an ActionRow that contains a single TextInput, used inside modals.
type TextInputRow struct {
	Type       int         `json:"type"` // always 1
	Components []TextInput `json:"components"`
}

// TextInput is a text field inside a modal (component type 4).
type TextInput struct {
	Type        int    `json:"type"` // always 4
	CustomID    string `json:"custom_id"`
	Style       int    `json:"style"` // 1 = short, 2 = paragraph
	Label       string `json:"label"`
	Required    bool   `json:"required,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	MaxLength   int    `json:"max_length,omitempty"`
}

// AutocompleteChoice is one option returned in an APPLICATION_COMMAND_AUTOCOMPLETE response.
type AutocompleteChoice struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// --- Incoming interaction payload ---

// InteractionMember holds the guild member who triggered the interaction.
type InteractionMember struct {
	User *struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
	Roles []string `json:"roles"`
}

// InteractionDataOption is a command option or subcommand in a slash command interaction.
type InteractionDataOption struct {
	Name    string                  `json:"name"`
	Type    int                     `json:"type"`
	Value   interface{}             `json:"value,omitempty"`
	Options []InteractionDataOption `json:"options,omitempty"`
	Focused bool                    `json:"focused,omitempty"`
}

// ModalTextInputValue is one submitted text-input value inside a MODAL_SUBMIT payload.
type ModalTextInputValue struct {
	Type     int    `json:"type"`
	CustomID string `json:"custom_id"`
	Value    string `json:"value"`
}

// ModalComponentRow is one row in the components array of a MODAL_SUBMIT payload.
type ModalComponentRow struct {
	Type       int                   `json:"type"`
	Components []ModalTextInputValue `json:"components,omitempty"`
}

// InteractionData is the command, component, or modal data attached to an Interaction.
type InteractionData struct {
	ID         string                  `json:"id,omitempty"`
	Name       string                  `json:"name,omitempty"`
	Options    []InteractionDataOption `json:"options,omitempty"`
	CustomID   string                  `json:"custom_id,omitempty"`
	Components []ModalComponentRow     `json:"components,omitempty"` // MODAL_SUBMIT values
}

// InteractionSourceMessage identifies the original message for a component interaction.
type InteractionSourceMessage struct {
	ID string `json:"id"`
}

// Interaction is the incoming webhook payload Discord sends for every user action.
type Interaction struct {
	ID            string                    `json:"id"`
	ApplicationID string                    `json:"application_id"`
	Type          int                       `json:"type"`
	GuildID       string                    `json:"guild_id"`
	Member        *InteractionMember        `json:"member"`
	Token         string                    `json:"token"`
	Data          *InteractionData          `json:"data"`
	Message       *InteractionSourceMessage `json:"message,omitempty"`
}

// interactionResponse is the JSON written back to Discord for every interaction.
type interactionResponse struct {
	Type int         `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// VerifyRequest validates the Ed25519 signature Discord attaches to every interaction webhook.
// publicKeyHex is the hex-encoded application public key from the Discord developer portal.
// Returns a non-nil error when the signature is absent or invalid.
func VerifyRequest(r *http.Request, body []byte, publicKeyHex string) error {
	sig := r.Header.Get("X-Signature-Ed25519")
	ts := r.Header.Get("X-Signature-Timestamp")
	if sig == "" || ts == "" {
		return errors.New("missing Discord signature headers")
	}
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return errors.New("invalid signature encoding")
	}
	pubBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return errors.New("invalid public key encoding")
	}
	if !ed25519.Verify(ed25519.PublicKey(pubBytes), append([]byte(ts), body...), sigBytes) {
		return errors.New("invalid request signature")
	}
	return nil
}

// buildEventEmbed constructs the RSVP announcement embed displayed in the events channel.
// maybeIDs and notAttendingIDs are only shown for Skirmish events.
func buildEventEmbed(eventType string, isQuick bool, description, hostID string, epochSecs int64, attendingIDs, maybeIDs, notAttendingIDs []string) Embed {
	attendeeText := "*No one yet — be the first!*"
	if len(attendingIDs) > 0 {
		parts := make([]string, len(attendingIDs))
		for i, id := range attendingIDs {
			parts[i] = "<@" + id + ">"
		}
		attendeeText = strings.Join(parts, "  ")
	}

	fields := []EmbedField{
		{Name: "Host", Value: "<@" + hostID + ">", Inline: true},
		{Name: "Starts", Value: "<t:" + strconv.FormatInt(epochSecs, 10) + ":F>", Inline: true},
	}
	if description != "" {
		fields = append(fields, EmbedField{Name: "📢 Rally Message", Value: description})
	}
	fields = append(fields, EmbedField{
		Name:  fmt.Sprintf("📋 Attending (%d)", len(attendingIDs)),
		Value: attendeeText,
	})

	color := 0x57F287 // green (quick event default)
	footerText := "Click Attending below to register"

	if !isQuick {
		color = 0xFEE75C // gold for large events
		footerText = "RSVP below — Maybe attendees get a 1-hour DM reminder"

		maybeText := "*No one yet*"
		if len(maybeIDs) > 0 {
			parts := make([]string, len(maybeIDs))
			for i, id := range maybeIDs {
				parts[i] = "<@" + id + ">"
			}
			maybeText = strings.Join(parts, "  ")
		}
		fields = append(fields, EmbedField{
			Name:  fmt.Sprintf("❓ Maybe (%d)", len(maybeIDs)),
			Value: maybeText,
		})
	}

	return Embed{
		Title:  "🎮 " + eventType,
		Color:  color,
		Fields: fields,
		Footer: &EmbedFooter{Text: footerText},
	}
}

// buildRSVPButtons returns the RSVP action row for an event embed.
// Quick events get 2 buttons; large events get 3 (adds Maybe).
func buildRSVPButtons(eventID string, isQuick bool) []ActionRow {
	buttons := []Component{
		{Type: 2, Style: 3, Label: "✅  Attending", CustomID: "event_join|" + eventID},
		{Type: 2, Style: 4, Label: "❌  Not Attending", CustomID: "event_decline|" + eventID},
	}
	if !isQuick {
		buttons = append(buttons, Component{
			Type: 2, Style: 2, Label: "❓  Maybe", CustomID: "event_maybe|" + eventID,
		})
	}
	return []ActionRow{{Type: 1, Components: buttons}}
}
