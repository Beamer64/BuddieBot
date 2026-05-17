package slash

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/bb_images/animated"
	"github.com/Beamer64/bb_images/color"
	"github.com/Beamer64/bb_images/edges"
	"github.com/Beamer64/bb_images/overlays"
	"github.com/Beamer64/bb_images/signs"
	"github.com/Beamer64/bb_images/spatial"
	"github.com/Beamer64/bb_images/special"
	"github.com/bwmarrin/discordgo"
)

func fetchImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch image: status %d", resp.StatusCode)
	}
	img, _, err := image.Decode(resp.Body)
	return img, err
}

func sendImgResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	// Defer immediately so heavy filters (Stringify, Triggered, etc.) get
	// Discord's 15-minute window instead of the 3-second initial-response
	// deadline that fires "Unknown interaction" 404s.
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction: %w", err)
	}

	client := cfg.Clients.Dagpi
	// Options[0] is the SubCommandGroup ("filter" | "distort" | "meme" | "template");
	// the actual effect subcommand and its args live one level deeper.
	options := i.ApplicationCommandData().Options[0].Options[0]

	var imgName string
	var bufferImage []byte
	var err error

	user, err := s.User(i.Member.User.ID)
	if err != nil {
		return err
	}

	// Resolve target user: prefer the first user-typed option among args,
	// otherwise default to the invoker. Handles text-only and text-first
	// commands (change-my-mind, tweet, youtube, ...) without overwriting
	// user to nil when Options[0] is a string.
	for _, opt := range options.Options {
		if opt.Type == discordgo.ApplicationCommandOptionUser {
			if u := opt.UserValue(s); u != nil {
				user = u
			}
			break
		}
	}
	errRespMsg := "Unable to edit image at this moment, please try later :("

	switch options.Name {
	case "pixelate":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = spatial.Pixelate(img, 8)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Pixelate.png"

	case "mirror":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = spatial.Mirror(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Mirror.png"

	case "flip-image":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = spatial.Flip(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "FlipImage.png"

	case "colors":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = spatial.Colors(img, 5)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Colors.png"

	case "murica":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = overlays.America(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "America.png"

	case "communism":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = overlays.Communism(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Communism.png"

	case "triggered":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = animated.Triggered(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Triggered.gif"

	case "expand":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = animated.Expand(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "ExpandImage.gif"

	case "wasted":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = overlays.Wasted(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Wasted.png"

	case "sketch":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = edges.Sketch(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Sketch.png"

	case "spin":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = animated.Spin(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "SpinImage.gif"

	case "petpet":
		bufferImage, err = client.PetPet(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "PetPet.png"

	case "bonk":
		bufferImage, err = client.Bonk(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Bonk.png"

	case "bomb":
		bufferImage, err = client.Bomb(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Bomb.png"

	case "shake":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = animated.Shake(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Shake.gif"

	case "invert":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = color.Invert(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Invert.png"

	case "sobel":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = edges.Sobel(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Sobel.png"

	case "hog":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = edges.Hog(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Hog.png"

	case "triangle":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = spatial.Triangle(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Triangle.png"

	case "blur":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = color.Blur(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Blur.png"

	case "rgb":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = special.RGB(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "RGB.png"

	case "angel":
		bufferImage, err = client.Angel(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Angel.png"

	case "satan":
		bufferImage, err = client.Satan(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Satan.png"

	case "delete":
		bufferImage, err = client.Delete(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Delete.png"

	case "fedora":
		bufferImage, err = client.Fedora(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Fedora.png"

	case "hitler":
		bufferImage, err = client.Hitler(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Hitler.png"

	case "lego":
		bufferImage, err = client.Lego(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Lego.png"

	case "wanted":
		bufferImage, err = client.Wanted(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Wanted.png"

	case "stringify":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = spatial.Stringify(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Stringify.png"

	case "burn":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = special.Burn(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Burn.png"

	case "earth":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = color.Earth(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Earth.png"

	case "freeze":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = color.Freeze(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Freeze.png"

	case "ground":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = color.Ground(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Ground.png"

	case "mosaic":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = spatial.Mosaic(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Mosaic.png"

	case "sithlord":
		bufferImage, err = client.Sithlord(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Sithlord.png"

	case "jail":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = overlays.Jail(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Jail.png"

	case "shatter":
		bufferImage, err = client.Shatter(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
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
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "pride.png"

	case "trash":
		bufferImage, err = client.Trash(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Trash.png"

	case "deepfry":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = color.Deepfry(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "deepfry.png"

	case "ascii":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = special.Ascii(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Ascii.png"

	case "charcoal":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = edges.Charcoal(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Charcoal.png"

	case "posterize":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = color.Posterize(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Posterize.png"

	case "sepia":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = color.Sepia(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Sepia.png"

	case "swirl":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = spatial.Swirl(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Swirl.png"

	case "paint":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = special.Paint(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Paint.png"

	case "night":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = color.Night(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "night.png"

	case "rainbow":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = animated.Rainbow(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Rainbow.gif"

	case "magik":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = spatial.Magik(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "Magik.png"

	case "5guys1girl":
		guy := options.Options[0].UserValue(s)
		girl := options.Options[1].UserValue(s)

		bufferImage, err = client.FivegOneg(guy.AvatarURL("300"), girl.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "fiveGuys.png"

	case "slap":
		slapped := options.Options[0].UserValue(s)
		slapper := options.Options[1].UserValue(s)

		bufferImage, err = client.Slap(slapper.AvatarURL("300"), slapped.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "slap.png"

	case "obama":
		bufferImage, err = client.Obama(user.AvatarURL("300"), user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
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

		avatar, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		displayName := user.GlobalName
		if displayName == "" {
			displayName = user.Username
		}
		bufferImage, err = signs.Tweet(avatar, displayName, user.Username, tweet)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
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

		avatar, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = signs.YouTube(avatar, user.Username, comment)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
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

		avatar, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		displayName := user.GlobalName
		if displayName == "" {
			displayName = user.Username
		}
		bufferImage, err = signs.Discord(avatar, displayName, msg)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "discord.png"

	case "retro-meme":
		// Both text args are optional now, so iterate by option name
		// instead of by position.
		var topText, bottomText string
		for _, opt := range options.Options {
			switch opt.Name {
			case "top-text":
				topText = opt.StringValue()
			case "bottom-text":
				bottomText = opt.StringValue()
			}
		}

		avatar, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = signs.RetroMeme(avatar, topText, bottomText)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "retro-meme.png"

	case "why_are_you_gay":
		user1 := options.Options[0].UserValue(s)
		user2 := options.Options[1].UserValue(s)

		bufferImage, err = client.WhyAreYouGay(user1.AvatarURL("300"), user2.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "why_are_you_gay.png"

	case "elmo":
		bufferImage, err = client.Elmo(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "elmo.png"

	case "tv-static":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = animated.TvStatic(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "static.gif"

	case "rain":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = animated.Rain(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "rain.gif"

	case "glitch":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = animated.Glitch(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "glitch.gif"

	case "sȶǟȶɨƈ-ɢʟɨȶƈɦ":
		img, fetchErr := fetchImage(user.AvatarURL("300"))
		if fetchErr != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fetchErr)
		}
		bufferImage, err = animated.GlitchStatic(img)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "static.gif"

	case "change-my-mind":
		text := options.Options[0].StringValue()

		bufferImage, err = signs.ChangeMyMind(text)
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "ChangeMyMind.png"

	case "album":
		bufferImage, err = client.Album(user.AvatarURL("300"))
		if err != nil {
			return helper.ReturnUserErrorDeferred(s, i, errRespMsg, err)
		}

		imgName = "album.png"

	}

	if _, err = s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{
				{
					Name:        imgName,
					ContentType: "image",
					Reader:      bytes.NewReader(bufferImage),
				},
			},
		},
	); err != nil {
		return fmt.Errorf("error editing interaction response: %w", err)
	}

	return nil
}

func imageSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "image",
		Description: "Image manipulation commands",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Name:        "filter",
				Description: "Color and tonal filters",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Name:        "blur",
						Description: "ig like pixelate?",
						Required:    false,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionUser,
								Name:        "user",
								Description: "Nobody should have to seem them",
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
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Name:        "distort",
				Description: "Pixel and geometric manipulation",
				Options: []*discordgo.ApplicationCommandOption{
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
						Name:        "mosaic",
						Description: "Turn an image into a roman mosaic",
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
						Name:        "shake",
						Description: "not stirred",
						Required:    false,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionUser,
								Name:        "user",
								Description: "shake them till they sleep",
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
						Name:        "triangle",
						Description: "shapes?",
						Required:    false,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionUser,
								Name:        "user",
								Description: "try my angle til I...rhombust..",
								Required:    false,
							},
						},
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Name:        "meme",
				Description: "Overlay a cultural element on the image",
				Options: []*discordgo.ApplicationCommandOption{
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
						Name:        "retro-meme",
						Description: "The good old memes. Generated.",
						Required:    false,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "top-text",
								Description: "top msg (optional)",
								Required:    false,
							},
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "bottom-text",
								Description: "bottom msg (optional)",
								Required:    false,
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
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Name:        "flag",
				Description: "represent",
				Options: []*discordgo.ApplicationCommandOption{
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
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Name:        "sign",
				Description: "Embed the image inside a framed meme",
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
						Name:        "change-my-mind",
						Description: "Prolly can't",
						Required:    false,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "text",
								Description: "unpopular opinion",
								Required:    true,
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
			},
		},
	}
}
