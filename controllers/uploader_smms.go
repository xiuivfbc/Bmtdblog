package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/wangsongyan/wblog/models"
	"github.com/wangsongyan/wblog/system"
)

type SmmsUploader struct {
}

type SmmsRet struct {
	Code    string `json:"code"`
	Success bool   `json:"success"`
	Data    struct {
		FileName  string `json:"filename"`
		StoreName string `json:"storename"`
		Size      int    `json:"size"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Hash      string `json:"hash"`
		Delete    string `json:"delete"`
		Url       string `json:"url"`
		Path      string `json:"path"`
		Msg       string `json:"msg"`
	} `json:"data"`
}

func (u SmmsUploader) upload(file multipart.File, fileHeader *multipart.FileHeader) (url string, err error) {
	var (
		resp      *http.Response
		req       *http.Request
		bodyBytes []byte
		ret       SmmsRet
		bodyBuf   = &bytes.Buffer{}
		smmsFile  models.SmmsFile
		cfg       = system.GetConfiguration()
	)
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("smfile", fileHeader.Filename)
	if err != nil {
		return
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return
	}
	err = bodyWriter.Close()
	if err != nil {
		return
	}

	req, err = http.NewRequest("POST", cfg.Smms.ApiUrl, bodyBuf)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	req.Header.Set("Authorization", cfg.Smms.ApiKey)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	bodyBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyBytes, &ret)
	if err != nil {
		return
	}
	if !ret.Success {
		err = errors.New(ret.Data.Msg)
		return
	}
	smmsFile = models.SmmsFile{
		FileName:  ret.Data.FileName,
		StoreName: ret.Data.StoreName,
		Size:      ret.Data.Size,
		Width:     ret.Data.Width,
		Height:    ret.Data.Height,
		Hash:      ret.Data.Hash,
		Delete:    ret.Data.Delete,
		Url:       ret.Data.Url,
		Path:      ret.Data.Path,
	}
	err = smmsFile.Insert()
	if err != nil {
		return
	}
	url = ret.Data.Url
	return
}
