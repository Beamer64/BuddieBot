package helper

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func GetErrorEmbed(err error, s *discordgo.Session, gID string) *discordgo.MessageEmbed {
	var guild *discordgo.Guild
	guildID := "N/A"
	guildName := "N/A"

	if gID != "" {
		guild, _ = s.Guild(gID)
		guildID = gID
		guildName = guild.Name
	}

	embed := &discordgo.MessageEmbed{
		Title:       "ERROR",
		Description: "(ノಠ益ಠ)ノ彡┻━┻",
		Color:       16726843,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Guild ID",
				Value:  guildID,
				Inline: true,
			},
			{
				Name:   "Guild Name",
				Value:  guildName,
				Inline: true,
			},
			{
				Name:   "Stack",
				Value:  fmt.Sprintf("%+v", errors.WithStack(err)),
				Inline: false,
			},
		},
	}

	return embed
}

func SendResponseErrorToUser(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6, // Ephemeral
				Content: message,
			},
		},
	)

	return err
}

// LogErrorsToErrorChannel logs any errors to console and send to the Error Discord Channel
func LogErrorsToErrorChannel(s *discordgo.Session, errorLogChannelID string, err error, guildID string) {
	fmt.Printf("%+v", errors.WithStack(err))
	_, _ = s.ChannelMessageSendEmbed(errorLogChannelID, GetErrorEmbed(err, s, guildID))
}

// IsLaunchedByDebugger Determines if application is being run by the debugger.
func IsLaunchedByDebugger() bool {
	// gops executable must be in the path. See https://github.com/google/gops
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	if err == nil && strings.Contains(string(gopsOut), "\\dlv.exe") {
		// our parent process is (probably) the Delve debugger
		return true
	}
	return false
}

func GetRandomStringFromSet(set []string) string {
	str := set[rand.Intn(len(set))]
	time.Sleep(1 * time.Millisecond)
	return str
}

func MemberHasRole(session *discordgo.Session, m *discordgo.Member, guildID string, roleName string) bool {
	if guildID == "" {
		guildID = m.GuildID
	}
	roleName = strings.ToLower(roleName)

	for _, roleID := range m.Roles {
		role, err := session.State.Role(guildID, roleID)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
		}

		if strings.ToLower(role.Name) == roleName {
			return true
		}
	}
	return false
}

// RangeIn Returns pseudo rand num between low and high.
// For random embed color: RangeIn(1, 16777215)
func RangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

// CheckIfStructValueISEmpty Checks if the value is empty and returns it as string if not.
// Otherwise, return 'N/A'
func CheckIfStructValueISEmpty(value interface{}) string {
	if value == nil {
		return "N/A"
	}

	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)

	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)

	case string:
		if v != "" && v != " " {
			return v
		}
		return "N/A"

	default:
		return "N/A"
	}
}

func StringInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func ToConvertedText(text string, convertGroup string) (string, error) {
	letters, err := getLetters()
	if err != nil {
		return "", err
	}

	convertedText := ""
	for _, char := range text {
		randSubs := ""
		subSet := letters[convertGroup][0][string(char)]
		if subSet != nil {
			randSubs = GetRandomStringFromSet(subSet)
		} else {
			randSubs = string(char)
		}
		convertedText += randSubs
	}

	return convertedText, nil
}

func getLetters() (map[string][]map[string][]string, error) {
	fontsDir := "datasets/text_fonts.json"
	if IsLaunchedByDebugger() {
		fontsDir = "../../datasets/text_fonts.json"
	}

	jsonFile, err := os.Open(fontsDir)
	if err != nil {
		return nil, err
	}

	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(jsonFile)

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	letters := make(map[string][]map[string][]string)
	err = json.Unmarshal(byteValue, &letters)
	if err != nil {
		return nil, err
	}

	return letters, nil
}

func CreateImgFile(imgPath string, img image.Image) error {
	// save imageData
	toimg, err := os.Create(imgPath)
	if err != nil {
		return err
	}
	fmt.Println("Created image location...")

	defer func(toimg *os.File) {
		err = toimg.Close()
	}(toimg)
	if err != nil {
		return err
	}

	err = png.Encode(toimg, img)
	if err != nil {
		return err
	}
	fmt.Println("Image Encoded...")

	return nil
}

func GetImgbbUploadURL(cfg *config.Configs, imgPath string, expireSecs ...int) (string, error) {
	apiUrl := fmt.Sprintf("%s&key=%s", cfg.Configs.ApiURLs.ImgbbAPI, cfg.Configs.Keys.ImgbbAPIkey)
	if expireSecs != nil {
		apiUrl = fmt.Sprintf("%sexpiration=%s&key=%s", cfg.Configs.ApiURLs.ImgbbAPI, strconv.Itoa(expireSecs[0]), cfg.Configs.Keys.ImgbbAPIkey)
	}

	// Read the entire file into a byte slice
	imgBytes, err := os.ReadFile(imgPath)
	if err != nil {
		return "", err
	}

	// Create a buffer to write out multipart form data to
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the imageData field
	base64img := base64.StdEncoding.EncodeToString(imgBytes)
	err = writer.WriteField("image", base64img)
	if err != nil {
		return "", fmt.Errorf("failed to write field: %w", err)
	}

	// Close the writer to finalize the form data
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", apiUrl, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var imgbbDataObject imgBBData
	err = json.Unmarshal(respBody, &imgbbDataObject)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return imgbbDataObject.Data.URL, nil
}
