package voiceChat

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
	"sync"
)

// Borrowed from https://github.com/bwmarrin/dgvoice

var (
	sendpcm bool
	// recv    chan *discordgo.Packet
	mu sync.Mutex
)

const (
	maxBytes = (frameSize * 2) * 2 // max size of opus data
)

func SendPCM(v *discordgo.VoiceConnection, pcm <-chan []int16) {
	mu.Lock()
	if sendpcm || pcm == nil {
		mu.Unlock()
		return
	}
	sendpcm = true
	mu.Unlock()
	defer func() { sendpcm = false }()

	opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder Error:", err)
		return
	}

	for {
		// read pcm from chan, exit if channel is closed.
		recv, ok := <-pcm
		if !ok {
			fmt.Println("PCM Channel closed.")
			return
		}

		// try encoding pcm frame with Opus
		opus, err := opusEncoder.Encode(recv, frameSize, maxBytes)
		if err != nil {
			fmt.Println("Encoding Error:", err)
			return
		}

		if v.Ready == false || v.OpusSend == nil {
			// fmt.Printf("Discordgo not ready for opus packets. %+v : %+v", v.Ready, v.OpusSend)
			return
		}
		// send encoded opus data to the sendOpus channel
		v.OpusSend <- opus
	}
}
