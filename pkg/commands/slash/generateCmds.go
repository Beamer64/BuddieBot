package slash

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/StephaneBunel/bresenham"
	"github.com/bwmarrin/discordgo"
	"github.com/chromedp/chromedp"
)

const (
	inputName string = "input"
	typeName  string = "type"
)

// generateOpts is the parsed-option map passed to each /generate validator.
type generateOpts = map[string]*discordgo.ApplicationCommandInteractionDataOption

// generateValidators dispatches per-cmdType input validation BEFORE the interaction is deferred.
// Adding a new /generate subcommand: write a validateXyz function near
// the subcommand's other code, add one entry here. cmdTypes without an
// entry are treated as having no validation requirements.
var generateValidators = map[string]func(*discordgo.Session, *discordgo.InteractionCreate, generateOpts) bool{
	"landsat":    validateLandsat,
	"cistercian": validateCistercian,
}

func sendGenerateResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	// Flat options: `type` (required choice), `text` (optional, required for some types).
	optMap := generateOpts{}
	for _, opt := range i.ApplicationCommandData().Options {
		optMap[opt.Name] = opt
	}
	cmdType := optMap[typeName].StringValue()
	errRespMsg := "Unable to make call at this moment, please try later :("

	if validate, ok := generateValidators[cmdType]; ok {
		if !validate(s, i, optMap) {
			return nil // validator already surfaced a user-facing message
		}
	}

	// Defer the interaction response to avoid timeout
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /generate %s %s: %w", typeName, cmdType, err)
	}

	var err error
	var embed *discordgo.MessageEmbed
	// Set by cases that need to attach a file alongside the embed; the
	// embed references the file via attachment:// — no third-party host.
	var (
		attachment     []byte
		attachmentName string
	)

	switch cmdType {
	case "landsat":
		var imgBytes []byte
		imgBytes, err = getLandsatImage(cfg, optMap[inputName].StringValue())
		if err == nil {
			attachment = imgBytes
			attachmentName = "landsat.png"
			embed = &discordgo.MessageEmbed{
				Title: "Landsat, more like...landFLAT...amirite non-round supporters??.",
				Color: helper.RandomDiscordColor(),
				Image: &discordgo.MessageEmbedImage{URL: "attachment://" + attachmentName},
				Footer: &discordgo.MessageEmbedFooter{
					Text: cfg.Configs.ApiURLs.LandsatAPI,
				},
			}
		}

	case "cistercian":
		// Validation above already confirmed parse + range
		n, _ := strconv.Atoi(strings.TrimSpace(optMap[inputName].StringValue()))

		negative := n < 0
		magnitude := n
		if negative {
			magnitude = -magnitude
		}
		var buf bytes.Buffer
		if err = png.Encode(&buf, drawCistLines(negative, fmt.Sprintf("%04d", magnitude))); err == nil {
			attachment = buf.Bytes()
			attachmentName = fmt.Sprintf("cistercian-%d.png", magnitude)
			embed = &discordgo.MessageEmbed{
				Title: fmt.Sprintf("Cistercian Numeral for %d", n),
				Color: helper.RandomDiscordColor(),
				Image: &discordgo.MessageEmbedImage{URL: "attachment://" + attachmentName},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "https://en.wikipedia.org/wiki/Cistercian_numerals",
				},
			}
		}

	case "fake-person":
		personData, err := callFakePersonAPI(cfg)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fmt.Errorf("fake-person API: %w", err))
		}

		embed = getFakePersonEmbed(personData)

	default:
		return fmt.Errorf("unknown option: %s", cmdType)
	}
	if err != nil {
		return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fmt.Errorf("sendGenerateResponse %s: %w", cmdType, err))
	}

	pingedUser := fmt.Sprintf("<@!%s>", i.Member.User.ID)
	webhookEdit := &discordgo.WebhookEdit{
		Content: &pingedUser,
		Embeds:  &[]*discordgo.MessageEmbed{embed},
	}
	if attachment != nil {
		webhookEdit.Files = []*discordgo.File{
			{
				Name:        attachmentName,
				ContentType: "image/png",
				Reader:      bytes.NewReader(attachment),
			},
		}
	}

	if _, err = s.InteractionResponseEdit(i.Interaction, webhookEdit); err != nil {
		return fmt.Errorf("send /generate response for %s %s: %w", typeName, cmdType, err)
	}
	return nil
}

func callFakePersonAPI(cfg *config.Configs) (fakePerson, error) {
	var personObj fakePerson

	resp, err := http.Get(cfg.Configs.ApiURLs.FakePersonAPI)
	if err != nil {
		return personObj, err
	}

	if resp.StatusCode != http.StatusOK {
		return personObj, fmt.Errorf("API call failed with status code %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&personObj)
	if err != nil {
		return personObj, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	return personObj, nil
}

func getFakePersonEmbed(fakePersonObj fakePerson) *discordgo.MessageEmbed {
	fpObj := fakePersonObj.Results[0]
	dob := strings.Split(fpObj.Dob.Date, "T")

	embed := &discordgo.MessageEmbed{
		Title:       "Fake Person Generator",
		Description: "BuddieBot has created life!",
		Color:       helper.RandomDiscordColor(),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Gender",
				Value:  fpObj.Gender,
				Inline: true,
			},
			{
				Name:   "Name",
				Value:  fmt.Sprintf("%s %s %s", fpObj.Name.Title, fpObj.Name.First, fpObj.Name.Last),
				Inline: true,
			},
			{
				Name:   "DOB",
				Value:  dob[0],
				Inline: true,
			},
			{
				Name:   "Age",
				Value:  fmt.Sprintf("%v", fpObj.Dob.Age),
				Inline: true,
			},
			{
				Name: "Address",
				Value: fmt.Sprintf(
					"%v %s\n%s, %s, %v %s", fpObj.Location.Street.Number, fpObj.Location.Street.Name, fpObj.Location.City, fpObj.Location.State, fpObj.Location.Postcode,
					fpObj.Location.Country,
				),
				Inline: false,
			},
			{
				Name:   "Email",
				Value:  fpObj.Email,
				Inline: true,
			},
			{
				Name:   "Username",
				Value:  fpObj.Login.Username,
				Inline: true,
			},
			{
				Name:   "Password",
				Value:  fpObj.Login.Password,
				Inline: true,
			},
			{
				Name:   "Phone",
				Value:  fpObj.Phone,
				Inline: true,
			},
			{
				Name:   "Cell",
				Value:  fpObj.Cell,
				Inline: true,
			},
			{
				Name:   fpObj.ID.Name,
				Value:  fpObj.ID.Value,
				Inline: true,
			},
			{
				Name:   "Nationality",
				Value:  fpObj.Nat,
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fpObj.Picture.Large,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Generated with randomuser.me",
		},
	}

	return embed
}

func validateCistercian(s *discordgo.Session, i *discordgo.InteractionCreate, opts generateOpts) bool {
	text := ""
	if opts[inputName] == nil || opts[inputName].StringValue() == "" {
		_ = helper.ReturnUserError(s, i, fmt.Sprintf("`%s` is required for /generate %s:Cistercian", inputName, typeName), nil)
		return false
	}

	if opts[inputName] != nil {
		text = strings.TrimSpace(opts[inputName].StringValue())
	}
	n, err := strconv.Atoi(text)
	if err != nil || n < helper.CistercianMin || n > helper.CistercianMax {
		_ = helper.ReturnUserError(s, i, fmt.Sprintf("Please enter a whole number from -9999 to 9999 for /generate %s:cistercian.", typeName), nil)
		return false
	}
	return true
}

// drawCistLines renders a cistercian-numeral glyph on a 200×200 canvas.
// digits MUST be exactly four decimal characters
// The stroke color is a single uniform random RGB per call so each
// generated glyph is visually distinct.
func drawCistLines(negative bool, digits string) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	col := color.RGBA{
		R: uint8(helper.RangeIn(0, 255)),
		G: uint8(helper.RangeIn(0, 255)),
		B: uint8(helper.RangeIn(0, 255)),
		A: 255,
	}

	if negative {
		bresenham.DrawLine(img, 60, 100, 140, 100, col)
	}
	bresenham.DrawLine(img, 100, 20, 100, 180, col)

	for pos, char := range digits {
		if char == '0' {
			continue // no glyph for zero — leave the quadrant blank
		}
		var x1, y1, x2, y2 int
		switch pos {
		case 0: // thousands (bottom-left)
			switch char {
			case '5':
				bresenham.DrawLine(img, 60, 180, 100, 140, col)
			case '7':
				bresenham.DrawLine(img, 60, 180, 60, 140, col)
			case '8':
				bresenham.DrawLine(img, 60, 140, 60, 180, col)
			case '9':
				bresenham.DrawLine(img, 60, 180, 60, 140, col)
				bresenham.DrawLine(img, 60, 140, 100, 140, col)
			}
			x1, y1, x2, y2 = helper.CistThous[string(char)].X1, helper.CistThous[string(char)].Y1, helper.CistThous[string(char)].X2, helper.CistThous[string(char)].Y2
		case 1: // hundreds (bottom-right)
			switch char {
			case '5':
				bresenham.DrawLine(img, 140, 180, 100, 140, col)
			case '7':
				bresenham.DrawLine(img, 140, 180, 140, 140, col)
			case '8':
				bresenham.DrawLine(img, 140, 140, 140, 180, col)
			case '9':
				bresenham.DrawLine(img, 140, 180, 140, 140, col)
				bresenham.DrawLine(img, 140, 140, 100, 140, col)
			}
			x1, y1, x2, y2 = helper.CistHunds[string(char)].X1, helper.CistHunds[string(char)].Y1, helper.CistHunds[string(char)].X2, helper.CistHunds[string(char)].Y2
		case 2: // tens (top-left)
			switch char {
			case '5':
				bresenham.DrawLine(img, 60, 20, 100, 60, col)
			case '7':
				bresenham.DrawLine(img, 60, 20, 60, 60, col)
			case '8':
				bresenham.DrawLine(img, 60, 60, 60, 20, col)
			case '9':
				bresenham.DrawLine(img, 60, 20, 60, 60, col)
				bresenham.DrawLine(img, 60, 60, 100, 60, col)
			}
			x1, y1, x2, y2 = helper.CistTens[string(char)].X1, helper.CistTens[string(char)].Y1, helper.CistTens[string(char)].X2, helper.CistTens[string(char)].Y2
		case 3: // ones (top-right)
			switch char {
			case '5':
				bresenham.DrawLine(img, 100, 60, 140, 20, col)
			case '7':
				bresenham.DrawLine(img, 140, 20, 140, 60, col)
			case '8':
				bresenham.DrawLine(img, 140, 60, 140, 20, col)
			case '9':
				bresenham.DrawLine(img, 140, 20, 140, 60, col)
				bresenham.DrawLine(img, 140, 60, 100, 60, col)
			}
			x1, y1, x2, y2 = helper.CistOnes[string(char)].X1, helper.CistOnes[string(char)].Y1, helper.CistOnes[string(char)].X2, helper.CistOnes[string(char)].Y2
		}
		bresenham.DrawLine(img, x1, y1, x2, y2, col)
	}
	return img
}

func validateLandsat(s *discordgo.Session, i *discordgo.InteractionCreate, opts generateOpts) bool {
	if opts[inputName] == nil || opts[inputName].StringValue() == "" {
		_ = helper.ReturnUserError(s, i, fmt.Sprintf("`%s` is required for /generate %s:landsat", inputName, typeName), nil)
		return false
	}
	if len(opts[inputName].StringValue()) > 20 {
		_ = helper.ReturnUserError(s, i, fmt.Sprintf("`%s` limit is 20 for /generate %s:landsat", inputName, typeName), nil)
		return false
	}
	if ok, retry := landsatLimiter.Allow(i.Member.User.ID); !ok {
		msg := fmt.Sprintf("Landsat is heavy — try again in `%.0fs`.", retry.Seconds())
		_ = helper.ReturnUserError(s, i, msg, nil)
		return false
	}
	return true
}

// landsatSem caps concurrent headless-Chrome instances. Each Chrome
// allocates ~200MB and lives ~5s
var landsatSem = make(chan struct{}, 2)

// landsatLimiter: per-user cooldown. Longer than the image limiter
// because each call costs ~10s of wall time and Chrome RAM
var landsatLimiter = helper.NewRateLimiter(30 * time.Second)

func getLandsatImage(cfg *config.Configs, text string) ([]byte, error) {
	landsatSem <- struct{}{}
	defer func() { <-landsatSem }()

	landsatUrl := cfg.Configs.ApiURLs.LandsatAPI

	ctx, cancel := chromedp.NewContext(context.Background())
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var buf []byte
	err := chromedp.Run(
		ctx,
		chromedp.Navigate(landsatUrl),
		chromedp.WaitVisible(`#nameInput`),
		chromedp.SendKeys(`#nameInput`, text, chromedp.NodeVisible),
		chromedp.WaitVisible(`#enterButton`),
		chromedp.Click(`#enterButton`),
		// Fixed sleep: no DOM signal observed so far reliably indicates
		// the JPG tiles have actually painted.
		chromedp.Sleep(5*time.Second),
		chromedp.Screenshot(`#nameBoxes`, &buf, chromedp.NodeVisible),
	)
	if err != nil {
		return nil, fmt.Errorf("landsat: chromedp run: %w", err)
	}

	return buf, nil
}

func generateSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "generate",
		Description: "Like regenerate, only without the \"re\"",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        typeName,
				Description: "Let's create something.",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "cistercian", Value: "cistercian"},
					{Name: "landsat", Value: "landsat"},
					{Name: "fake-person", Value: "fake-person"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        inputName,
				Description: "Input",
				Required:    false,
			},
		},
	}
}
