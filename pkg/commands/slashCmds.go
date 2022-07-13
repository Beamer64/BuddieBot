package commands

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var (

	// Commands All commands and options must have a description
	// Commands/options without description will fail the registration
	// of the command.
	Commands = []*discordgo.ApplicationCommand{
		{
			Name:        "animals",
			Description: "So CUTE",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "doggo",
					Description: "🐕",
					Required:    false,
				},
			},
		},
		{
			Name:        "ratethis",
			Description: "Rate this ...",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "simp",
					Description: "Simp Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User simp score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "epicgamer",
					Description: "Epic Gamer Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Epic Gamer score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "dank",
					Description: "Dank Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Dank score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "gay",
					Description: "Gay Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Gay score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "schmeat",
					Description: "Schmeat Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Schmeat score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "stinky",
					Description: "Stinky Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Stinky score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "thot",
					Description: "Thot Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Thot score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "neckbeard",
					Description: "Neck Beard Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Neck Beard score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "pickme",
					Description: "Pick Me Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Pick Me score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "looks",
					Description: "Looks Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Looks score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "smarts",
					Description: "Smarts Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Smarts score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "nerd",
					Description: "Nerd Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Nerd score",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "geek",
					Description: "Geek Rating",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User Geek score",
							Required:    false,
						},
					},
				},
			},
		},
		{
			Name:        "get",
			Description: "Get a text based response like a joke or pickup line",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "rekd",
					Description: "Insult someone",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "nerd",
							Description: "Someone to insult",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "joke",
					Description: "Tell a joke",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "8ball",
					Description: "Think of a question",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "yomomma",
					Description: "is sooooooo fat..",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "Somones momma",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "pickup-line",
					Description: "Woah Momma",
					Required:    false,
				},
				/*{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "captcha",
					Description: "Are you a robot?",
					Required:    false,
				},*/
			},
		},
		{
			Name:        "imgset1",
			Description: "Manipulate some images!",
			Options: []*discordgo.ApplicationCommandOption{
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
					Name:        "shake",
					Description: "Shake a gif by having it wiggle",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "Shake someone else",
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
					Name:        "triangle",
					Description: "Cool triangle effect for an image",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "Triangle someone else",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "blur",
					Description: "Blurs a given image",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "Blur someone else",
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
			},
		},
		{
			Name:        "imgset2",
			Description: "Manipulate some more images!",
			Options: []*discordgo.ApplicationCommandOption{
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
			},
		},
		{
			Name:        "imgset3",
			Description: "MOAR!",
			Options: []*discordgo.ApplicationCommandOption{
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
					Name:        "sȶǟȶɨƈ-ɢʟɨȶƈɦ",
					Description: "ʟɨʄɛ ɨֆ քǟɨռ",
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
			},
		},
		{
			Name:        "daily",
			Description: "Receive daily quotes, horoscopes, affirmations, etc.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "horoscope",
					Description: "Gives daily horoscope",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "affirmation",
					Description: "Gives daily affirmation",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "kanye",
					Description: "Gifts us with a quote from the man himself",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "advice",
					Description: "Words of wisdom",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "fact",
					Description: "Read a fun fact",
					Required:    false,
				},
			},
		},
		{
			Name:        "pick",
			Description: "I'll pick stuff for you. I'll also pick a steam game with the 1st choice of 'steam'",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "choices",
					Description: "Will choose between 2 or more things.",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "1st",
							Description: "First choice",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "2nd",
							Description: "Second choice",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "3rd",
							Description: "Third choice",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "4th",
							Description: "Fourth choice",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "5th",
							Description: "Fifth choice",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "6th",
							Description: "Sixth choice",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "album",
					Description: "I can recommend an album for you to listen to!",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "1st",
							Description: "First tag",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "2nd",
							Description: "Second tag",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "3rd",
							Description: "Third tag",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "4th",
							Description: "Fourth tag",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "5th",
							Description: "Fifth tag",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "6th",
							Description: "Sixth tag",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "steam",
					Description: "Will choose a random Steam game to play.",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "poll",
					Description: "Gauge the room!",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "request",
							Description: "Post the Question",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "1st_poll_item",
							Description: "First Choice",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "2nd_poll_item",
							Description: "Second Choice",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "3rd_poll_item",
							Description: "Third Choice",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "4th_poll_item",
							Description: "Fourth Choice",
							Required:    false,
						},
					},
				},
			},
		},
		{
			Name:        "play",
			Description: "Play some games! *More coming soon",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "coin-flip",
					Description: "Flips a coin...",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "wyr",
					Description: "Would You Rather??",
					Required:    false,
				},
				/*{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "nim",
					Description: "of the 12 coin variety",
					Required:    false,
				},*/
				/*{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "typeracer",
					Description: "It's like a game or something.",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "gtl",
					Description: "Guess that Logo!",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "wtp",
					Description: "Who's that Pokemon?!",
					Required:    false,
				},*/
			},
		},
		{
			Name:        "txt",
			Description: "Funky Texts",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "clapback",
					Description: "Say it with sass",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "text",
							Description: "Text to change",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "bubble",
					Description: "Bubble Text",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "text",
							Description: "Text to change",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "tuuck",
			Description: "I've fallen and can't get up!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "Specify a command for a description",
					Required:    false,
				},
			},
		},
		{
			Name:        "config",
			Description: "set guild settings",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "list settings",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "setting",
					Description: "change specific setting",
					Required:    false,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "name of setting",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "new value for setting",
							Required:    true,
						},
					},
				},
			},
		},
	}

	// ComponentHandlers for handling components in interactions. Eg. Buttons, Dropdowns, Searchbars Etc.
	ComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs){
		"horo-select": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendHoroscopeCompResponse(s, i)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"album-suggest": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendAlbumPickCompResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"wyr-button": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendWYRCompResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},
	}

	// CommandHandlers for handling the commands themselves. Main interaction response here.
	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs){
		"animals": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendAnimalsResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"txt": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendTxtResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"ratethis": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendRateThisResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"get": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendGetResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"img-set1": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendImgResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"img-set2": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendImgResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"img-set3": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendImgResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"daily": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendDailyResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"pick": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendPickResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"tuuck": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendTuuckResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendPlayResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},

		"config": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendConfigResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, i.GuildID))
			}
		},
	}
)
