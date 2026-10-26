package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

type Format struct {
	Filename       string            `json:"filename"`
	FormatName     string            `json:"format_name"`
	FormatLongName string            `json:"format_long_name"`
	StartTime      float64           `json:"start_time,string"`
	Duration       float64           `json:"duration,string"`
	Size           int64             `json:"size,string"`
	BitRate        int               `json:"bit_rate,string"`
	TagList        map[string]string `json:"tags,omitempty"`
}

// Stream represents an individual track (video, audio, subtitle).
type Stream struct {
	Index         int               `json:"index"`
	CodecName     string            `json:"codec_name"`
	CodecType     string            `json:"codec_type"`
	CodecLongName string            `json:"codec_long_name"`
	Width         int               `json:"width,omitempty"`
	Height        int               `json:"height,omitempty"`
	RFrameRate    string            `json:"r_frame_rate,omitempty"`
	AvgFrameRate  string            `json:"avg_frame_rate,omitempty"`
	BitRate       string            `json:"bit_rate,omitempty"`
	Duration      string            `json:"duration,omitempty"`
	SampleRate    string            `json:"sample_rate,omitempty"`
	Channels      int               `json:"channels,omitempty"`
	ChannelLayout string            `json:"channel_layout,omitempty"`
	TagList       map[string]string `json:"tags,omitempty"`
}
type ProbeData struct {
	Format  Format   `json:"format"`
	Streams []Stream `json:"streams"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	buf := bytes.Buffer{}
	cmd.Stdout = &buf
	err := cmd.Run()

	if err != nil {
		return "", err
	}

	var jsonOut ProbeData
	err = json.Unmarshal(buf.Bytes(), &jsonOut)
	if err != nil {
		return "", err
	}

	width := float64(jsonOut.Streams[0].Width)
	height := float64(jsonOut.Streams[0].Height)
	ratio := width / height

	if ratio >= 1.7 && ratio <= 1.8 {
		return "16:9", nil
	}

	if ratio >= 0.56 && ratio <= 0.57 {
		return "9:16", nil
	}

	return "other", nil

}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
