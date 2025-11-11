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
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Pixelate.png"

	case "mirror":
		bufferImage, err = client.Mirror(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Mirror.png"

	case "flip-image":
		bufferImage, err = client.FlipImage(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "FlipImage.png"

	case "colors":
		bufferImage, err = client.Colors(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Colors.png"

	case "murica":
		bufferImage, err = client.America(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "America.png"

	case "communism":
		bufferImage, err = client.Communism(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Communism.png"

	case "triggered":
		bufferImage, err = client.Triggered(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Triggered.png"

	case "expand":
		bufferImage, err = client.ExpandImage(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "ExpandImage.png"

	case "wasted":
		bufferImage, err = client.Wasted(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Wasted.png"

	case "sketch":
		bufferImage, err = client.Sketch(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Sketch.png"

	case "spin":
		bufferImage, err = client.SpinImage(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "SpinImage.png"

	case "petpet":
		bufferImage, err = client.PetPet(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "PetPet.png"

	case "bonk":
		bufferImage, err = client.Bonk(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Bonk.png"

	case "bomb":
		bufferImage, err = client.Bomb(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Bomb.png"

	case "shake":
		bufferImage, err = client.Shake(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Shake.png"

	case "invert":
		bufferImage, err = client.Invert(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Invert.png"

	case "sobel":
		bufferImage, err = client.Sobel(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Sobel.png"

	case "hog":
		bufferImage, err = client.Hog(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Hog.png"

	case "triangle":
		bufferImage, err = client.Triangle(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Triangle.png"

	case "blur":
		bufferImage, err = client.Blur(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Blur.png"

	case "rgb":
		bufferImage, err = client.RGB(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "RGB.png"

	case "angel":
		bufferImage, err = client.Angel(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Angel.png"

	case "satan":
		bufferImage, err = client.Satan(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Satan.png"

	case "delete":
		bufferImage, err = client.Delete(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Delete.png"

	case "fedora":
		bufferImage, err = client.Fedora(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Fedora.png"

	case "hitler":
		bufferImage, err = client.Hitler(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Hitler.png"

	case "lego":
		bufferImage, err = client.Lego(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Lego.png"

	case "wanted":
		bufferImage, err = client.Wanted(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Wanted.png"

	case "stringify":
		bufferImage, err = client.Stringify(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Stringify.png"

	case "burn":
		bufferImage, err = client.Burn(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Burn.png"

	case "earth":
		bufferImage, err = client.Earth(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Earth.png"

	case "freeze":
		bufferImage, err = client.Freeze(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Freeze.png"

	case "ground":
		bufferImage, err = client.Ground(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Ground.png"

	case "mosiac":
		bufferImage, err = client.Mosiac(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Mosiac.png"

	case "sithlord":
		bufferImage, err = client.Sithlord(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Sithlord.png"

	case "jail":
		bufferImage, err = client.Jail(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Jail.png"

	case "shatter":
		bufferImage, err = client.Shatter(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
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
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "pride.png"

	case "trash":
		bufferImage, err = client.Trash(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Trash.png"

	case "deepfry":
		bufferImage, err = client.Deepfry(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "deepfry.png"

	case "ascii":
		bufferImage, err = client.Ascii(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Ascii.png"

	case "charcoal":
		bufferImage, err = client.Charcoal(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Charcoal.png"

	case "posterize":
		bufferImage, err = client.Posterize(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Posterize.png"

	case "sepia":
		bufferImage, err = client.Sepia(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Sepia.png"

	case "swirl":
		bufferImage, err = client.Swirl(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Swirl.png"

	case "paint":
		bufferImage, err = client.Paint(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Paint.png"

	case "night":
		bufferImage, err = client.Night(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "night.png"

	case "rainbow":
		bufferImage, err = client.Rainbow(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Rainbow.png"

	case "magik":
		bufferImage, err = client.Magik(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Magik.png"

	case "5guys1girl":
		guy := options.Options[0].UserValue(s)
		girl := options.Options[1].UserValue(s)

		bufferImage, err = client.FivegOneg(guy.AvatarURL("300"), girl.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "fiveGuys.png"

	case "slap":
		slapped := options.Options[0].UserValue(s)
		slapper := options.Options[1].UserValue(s)

		bufferImage, err = client.Slap(slapper.AvatarURL("300"), slapped.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "slap.png"

	case "obama":
		bufferImage, err = client.Obama(user.AvatarURL("300"), user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
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
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
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
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
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
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
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
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
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
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
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
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "modern-meme.png"

	case "why_are_you_gay":
		user1 := options.Options[0].UserValue(s)
		user2 := options.Options[1].UserValue(s)

		bufferImage, err = client.WhyAreYouGay(user1.AvatarURL("300"), user2.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "why_are_you_gay.png"

	case "elmo":
		bufferImage, err = client.Elmo(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "elmo.png"

	case "tv-static":
		bufferImage, err = client.TvStatic(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "static.png"

	case "rain":
		bufferImage, err = client.Rain(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "rain.png"

	case "glitch":
		bufferImage, err = client.Glitch(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "glitch.png"

	case "sȶǟȶɨƈ-ɢʟɨȶƈɦ":
		bufferImage, err = client.GlitchStatic(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "static.png"

	case "album":
		bufferImage, err = client.Album(user.AvatarURL("300"))
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
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
		return fmt.Errorf("error sendind Interaction: %v", err)
	}

	return nil
}
