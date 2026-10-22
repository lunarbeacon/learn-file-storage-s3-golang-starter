package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	// set the maximum upload size limit on the request body
	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)

	// get the video ID from the request url and parse it to a uuid
	videoIDstring := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDstring)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Video ID", err)
		return
	}

	// validate the user and get their user ID
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	_, err = auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	// get the video metadata from the database
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find video", err)
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid file", err)
	}
	defer file.Close()

	mediatype, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse media type", err)
		return
	}

	if mediatype != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Unsupported media type", err)
		return
	}

	tmpFile, err := os.CreateTemp("", "tubely-upload.mp4")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to create temp file", err)
		return
	}

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to create copy temp file", err)
	}

	tmpFile.Seek(0, io.SeekStart)

	// name the file we are placing on s3 bucket
	randName := make([]byte, 32)
	_, err = rand.Read(randName)
	s3FileName := base64.RawURLEncoding.EncodeToString(randName) + ".mp4"

	put := s3.PutObjectInput{}
	put.Bucket = &cfg.s3Bucket
	put.Key = &s3FileName
	put.Body = tmpFile
	put.ContentType = &mediatype

	// upload it
	_, err = cfg.s3Client.PutObject(context.Background(), &put)

	videoURL := "http://tubely-11066.s3." + cfg.s3Region + ".amazonaws.com/" + s3FileName
	formatted := time.Now().Format("2006-01-02 15:04:05")
	updateTime, _ := time.Parse("2006-01-02 15:04:05", formatted)
	video.UpdatedAt = updateTime
	video.VideoURL = &videoURL

	fmt.Printf("video: %v\n", video)

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to upload video to s3", err)
		return
	}

	return

}
