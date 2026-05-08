package slash

import (
	"bytes"
	"fmt"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func sendImgResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	client := cfg.Clients.Dagpi
	options := i.ApplicationCommandData().Options[0]

	var imgName string
	var bufferImage []byte
	var err error

	user, err := s.User(i.Member.User.ID)
	if err != nil {
		return err
	}

	if len(options.Options) > 0 {
		user = options.Options[0].UserValue(s)
	}
	errRespMsg := "Unable to edit image at this moment, please try later :("

	switch options.Name {
	case "pixelate":
		bufferImage, err = client.Pixelate(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Pixelate.png"

	case "mirror":
		bufferImage, err = client.Mirror(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Mirror.png"

	case "flip-image":
		bufferImage, err = client.FlipImage(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "FlipImage.png"

	case "colors":
		bufferImage, err = client.Colors(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Colors.png"

	case "murica":
		bufferImage, err = client.America(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "America.png"

	case "communism":
		bufferImage, err = client.Communism(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Communism.png"

	case "triggered":
		bufferImage, err = client.Triggered(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Triggered.png"

	case "expand":
		bufferImage, err = client.ExpandImage(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "ExpandImage.png"

	case "wasted":
		bufferImage, err = client.Wasted(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Wasted.png"

	case "sketch":
		bufferImage, err = client.Sketch(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Sketch.png"

	case "spin":
		bufferImage, err = client.SpinImage(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "SpinImage.png"

	case "petpet":
		bufferImage, err = client.PetPet(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "PetPet.png"

	case "bonk":
		bufferImage, err = client.Bonk(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Bonk.png"

	case "bomb":
		bufferImage, err = client.Bomb(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Bomb.png"

	case "shake":
		bufferImage, err = client.Shake(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Shake.png"

	case "invert":
		bufferImage, err = client.Invert(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Invert.png"

	case "sobel":
		bufferImage, err = client.Sobel(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Sobel.png"

	case "hog":
		bufferImage, err = client.Hog(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Hog.png"

	case "triangle":
		bufferImage, err = client.Triangle(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Triangle.png"

	case "blur":
		bufferImage, err = client.Blur(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Blur.png"

	case "rgb":
		bufferImage, err = client.RGB(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "RGB.png"

	case "angel":
		bufferImage, err = client.Angel(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Angel.png"

	case "satan":
		bufferImage, err = client.Satan(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Satan.png"

	case "delete":
		bufferImage, err = client.Delete(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Delete.png"

	case "fedora":
		bufferImage, err = client.Fedora(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Fedora.png"

	case "hitler":
		bufferImage, err = client.Hitler(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Hitler.png"

	case "lego":
		bufferImage, err = client.Lego(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Lego.png"

	case "wanted":
		bufferImage, err = client.Wanted(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Wanted.png"

	case "stringify":
		bufferImage, err = client.Stringify(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Stringify.png"

	case "burn":
		bufferImage, err = client.Burn(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Burn.png"

	case "earth":
		bufferImage, err = client.Earth(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Earth.png"

	case "freeze":
		bufferImage, err = client.Freeze(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Freeze.png"

	case "ground":
		bufferImage, err = client.Ground(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Ground.png"

	case "mosiac":
		bufferImage, err = client.Mosiac(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Mosiac.png"

	case "sithlord":
		bufferImage, err = client.Sithlord(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Sithlord.png"

	case "jail":
		bufferImage, err = client.Jail(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Jail.png"

	case "shatter":
		bufferImage, err = client.Shatter(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Shatter.png"

	case "pride":
		flag := options.Options[0].StringValue()

		switch len(options.Options) {
		case 1:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 2:
			user = options.Options[1].UserValue(s)
		}

		bufferImage, err = client.Pride(user.AvatarURL("300"), flag)
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "pride.png"

	case "trash":
		bufferImage, err = client.Trash(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Trash.png"

	case "deepfry":
		bufferImage, err = client.Deepfry(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "deepfry.png"

	case "ascii":
		bufferImage, err = client.Ascii(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Ascii.png"

	case "charcoal":
		bufferImage, err = client.Charcoal(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Charcoal.png"

	case "posterize":
		bufferImage, err = client.Posterize(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Posterize.png"

	case "sepia":
		bufferImage, err = client.Sepia(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Sepia.png"

	case "swirl":
		bufferImage, err = client.Swirl(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Swirl.png"

	case "paint":
		bufferImage, err = client.Paint(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Paint.png"

	case "night":
		bufferImage, err = client.Night(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "night.png"

	case "rainbow":
		bufferImage, err = client.Rainbow(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Rainbow.png"

	case "magik":
		bufferImage, err = client.Magik(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "Magik.png"

	case "5guys1girl":
		guy := options.Options[0].UserValue(s)
		girl := options.Options[1].UserValue(s)

		bufferImage, err = client.FivegOneg(guy.AvatarURL("300"), girl.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "fiveGuys.png"

	case "slap":
		slapped := options.Options[0].UserValue(s)
		slapper := options.Options[1].UserValue(s)

		bufferImage, err = client.Slap(slapper.AvatarURL("300"), slapped.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "slap.png"

	case "obama":
		bufferImage, err = client.Obama(user.AvatarURL("300"), user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "obama.png"

	case "tweet":
		tweet := options.Options[0].StringValue()

		switch len(options.Options) {
		case 1:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 2:
			user = options.Options[1].UserValue(s)
		}

		bufferImage, err = client.Tweet(user.AvatarURL("300"), user.Username, tweet)
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "tweet.png"

	case "youtube":
		comment := options.Options[0].StringValue()
		switch len(options.Options) {
		case 1:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 2:
			user = options.Options[1].UserValue(s)
		}

		bufferImage, err = client.YouTubeComment(user.AvatarURL("300"), user.Username, comment, false)
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "youtube.png"

	case "discord":
		msg := options.Options[0].StringValue()
		switch len(options.Options) {
		case 1:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 2:
			user = options.Options[1].UserValue(s)
		}

		bufferImage, err = client.Discord(user.AvatarURL("300"), user.Username, msg, true)
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "discord.png"

	case "retro-meme":
		topText := options.Options[0].StringValue()
		bottomText := options.Options[1].StringValue()

		switch len(options.Options) {
		case 2:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 3:
			user = options.Options[2].UserValue(s)
		}

		bufferImage, err = client.Retromeme(user.AvatarURL("300"), topText, bottomText)
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "retro-meme.png"

	case "motivational":
		topText := options.Options[0].StringValue()
		bottomText := options.Options[1].StringValue()

		switch len(options.Options) {
		case 2:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 3:
			user = options.Options[2].UserValue(s)
		}

		bufferImage, err = client.Motivational(user.AvatarURL("300"), topText, bottomText)
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "motivational.png"

	case "modern-meme":
		text := options.Options[0].StringValue()

		switch len(options.Options) {
		case 1:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 2:
			user = options.Options[1].UserValue(s)
		}

		bufferImage, err = client.Modernmeme(user.AvatarURL("300"), text)
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "modern-meme.png"

	case "why_are_you_gay":
		user1 := options.Options[0].UserValue(s)
		user2 := options.Options[1].UserValue(s)

		bufferImage, err = client.WhyAreYouGay(user1.AvatarURL("300"), user2.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "why_are_you_gay.png"

	case "elmo":
		bufferImage, err = client.Elmo(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "elmo.png"

	case "tv-static":
		bufferImage, err = client.TvStatic(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "static.png"

	case "rain":
		bufferImage, err = client.Rain(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "rain.png"

	case "glitch":
		bufferImage, err = client.Glitch(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "glitch.png"

	case "sȶǟȶɨƈ-ɢʟɨȶƈɦ":
		bufferImage, err = client.GlitchStatic(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "static.png"

	case "album":
		bufferImage, err = client.Album(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserError(s, i, errRespMsg, err)
		}

		imgName = "album.png"

	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Files: []*discordgo.File{
					{
						Name:        imgName,
						ContentType: "image",
						Reader:      bytes.NewReader(bufferImage),
					},
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("error sendind Interaction: %w", err)
	}

	return nil
}

func imgBSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "img-b",
		Description: "Manipulate some images!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "album",
				Description: "Make an album cover!",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "angel",
				Description: "Image on the Angels face",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Make someone else an angel",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "ascii",
				Description: "Cool hackerman effect for an image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "bomb",
				Description: "Cool guys don't look at explosions",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Explode someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "bonk",
				Description: "Get bonked on my cheems",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Bonk someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "burn",
				Description: "Light your image on fire",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "charcoal",
				Description: "mage into a charcoal drawing",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "colors",
				Description: "Get an Image with the colors present in the image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Colors someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "communism",
				Description: "Support the soviet union comrade. Let the red flag fly!",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Convert someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "delete",
				Description: "Generates a windows error meme based on a given image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Delete someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "deepfry",
				Description: "Deepfry an image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "discord",
				Description: "Generate realistic discord messages",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "message",
						Description: "Message to be displayed",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Name and image to be displayed",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "earth",
				Description: "The green and blue of the earth",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "elmo",
				Description: "Burning Elmo meme 🔥🔥",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "expand",
				Description: "Animation that stretches an image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Expand someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "fedora",
				Description: "Tips fedora in appreciation. *Platypus noise*.",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Fedora someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "flip-image",
				Description: "Flip an image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Flip someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "freeze",
				Description: "Blue ice like tint",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "glitch",
				Description: "Are you there, Neo?",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "ground",
				Description: "The power of the earth",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "hitler",
				Description: "?????",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "hog",
				Description: "Histogram of Oriented Gradients for an image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Histogram someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "invert",
				Description: "Get an image with an inverted color effect",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Invert someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "jail",
				Description: "Put an image behind bars",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "lego",
				Description: "Every group of pixels is a lego brick",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
		},
	}
}

func imgESpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "img-e",
		Description: "Manipulate some more images!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "magik",
				Description: "The much loved magik endpoint",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "mirror",
				Description: "Mirror an image along the y-axis",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Mirror someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "modern-meme",
				Description: "A modern meme generation system that allows reddit ready memes with just one endpoint",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "text",
						Description: "it's top the text",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "mosiac",
				Description: "Turn an image into a roman mosiac",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "motivational",
				Description: "The black background with top and bottom motivational text.",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "top-text",
						Description: "top msg",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "bottom-text",
						Description: "bottom msg",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "murica",
				Description: "Let the star spangled banner of the free and the brave soar",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Murica someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "night",
				Description: "Turn an day into night",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "obama",
				Description: "What's his last name?!",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "ApplicationCommandOptionUser",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "paint",
				Description: "Turn an image into art",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "petpet",
				Description: "Pet pet",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Pet someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "pixelate",
				Description: "Pixelate yourself",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Mirror someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "posterize",
				Description: "Posterizes an image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "pride",
				Description: "Flag of your choice over an Image!",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "flag",
						Description: "Choose a flag",
						Required:    true,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{
								Name:  "asexual",
								Value: "asexual",
							},
							{
								Name:  "bisexual",
								Value: "bisexual",
							},
							{
								Name:  "gay",
								Value: "gay",
							},
							{
								Name:  "genderfluid",
								Value: "genderfluid",
							},
							{
								Name:  "genderqueer",
								Value: "genderqueer",
							},
							{
								Name:  "intersex",
								Value: "intersex",
							},
							{
								Name:  "lesbian",
								Value: "lesbian",
							},
							{
								Name:  "nonbinary",
								Value: "nonbinary",
							},
							{
								Name:  "progress",
								Value: "progress",
							},
							{
								Name:  "pan",
								Value: "pan",
							},
							{
								Name:  "trans",
								Value: "trans",
							},
						},
					},
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Flag someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "rain",
				Description: "For the rainy days",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "rainbow",
				Description: "Some trippy light effects",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "retro-meme",
				Description: "The good old memes. Generated.",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "top-text",
						Description: "top msg",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "bottom-text",
						Description: "bottom msg",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "rgb",
				Description: "Get an RGB graph of an image's colors",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "RGB someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "satan",
				Description: "Put an image on the devil 😈",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Make someone else the devil",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "sepia",
				Description: "Sepia Tone an image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "shatter",
				Description: "Broken glass overlay",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "sithlord",
				Description: "Put an image on the Laughs in Sithlord meme",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "sketch",
				Description: "Cool effect that shows how an image would have been created by an artist",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Sketch someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "slap",
				Description: "Have one image slap another",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "slapper",
						Description: "user",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "slapped",
						Description: "user",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "sobel",
				Description: "Get an image with the sobel effect",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Sobel someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "spin",
				Description: "You spin me right round baby",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Spin someone else",
						Required:    false,
					},
				},
			},
		},
	}
}

func imgWbsSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "img-wbs",
		Description: "MOAR!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "5guys1girl",
				Description: "The meme",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "guys",
						Description: "user",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "girl",
						Description: "user",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "stringify",
				Description: "Turn your image into a ball of yarn",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "swirl",
				Description: "Swirl an image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "trash",
				Description: "Image is trash",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "triggered",
				Description: "Get a triggered gif",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Trigger someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "tweet",
				Description: "Cast out to the void!",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "tweet",
						Description: "Message to be displayed",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Name and image to be displayed",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "tv-static",
				Description: "Tastes like Monster Energy™️",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "wanted",
				Description: "Wanted poster of an image",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "wasted",
				Description: "Get an image with GTA V Wasted screen",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Waste someone else",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "why_are_you_gay",
				Description: "The meme",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "interviewer",
						Description: "The interviewer",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "gay",
						Description: "The gay",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "youtube",
				Description: "Generate realistic Youtube messages",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "comment",
						Description: "Message to be displayed",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Name and image to be displayed",
						Required:    false,
					},
				},
			},
		},
	}
}
