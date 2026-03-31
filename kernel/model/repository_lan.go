package model

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/88250/gulu"
	"github.com/imroc/req/v3"
	"github.com/siyuan-note/dejavu/cloud"
	"github.com/siyuan-note/dejavu/entity"
)

type lanCloud struct {
	*cloud.BaseCloud
	client *req.Client
}

func newLANCloud(baseCloud *cloud.BaseCloud) *lanCloud {
	ensureLANConf()
	client := req.C()
	client.SetTimeout(time.Duration(Conf.Sync.LAN.Timeout) * time.Second)
	return &lanCloud{
		BaseCloud: baseCloud,
		client:    client,
	}
}

func (c *lanCloud) endpoint(p string) string {
	return strings.TrimRight(Conf.Sync.LAN.Endpoint, "/") + p
}

func (c *lanCloud) newRequest() *req.Request {
	request := c.client.R()
	request.SetHeader("X-Padhivu-LAN-Token", Conf.Sync.LAN.AuthToken)
	request.SetHeader("X-Padhivu-LAN-Device-ID", Conf.Sync.LAN.DeviceID)
	request.SetHeader("X-Padhivu-LAN-Device-Name", Conf.Sync.LAN.DeviceName)
	request.SetHeader("X-Padhivu-LAN-Device-OS", Conf.System.OS)
	return request
}

func (c *lanCloud) call(path string, payload map[string]interface{}, result interface{}) error {
	var envelope struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data"`
	}
	resp, err := c.newRequest().SetBodyJsonMarshal(payload).SetSuccessResult(&envelope).Post(c.endpoint(path))
	if err != nil {
		return err
	}
	if !resp.IsSuccessState() {
		return fmt.Errorf("LAN sync request failed [%d]", resp.StatusCode)
	}
	if 0 != envelope.Code {
		if "" != envelope.Msg {
			return errors.New(envelope.Msg)
		}
		return fmt.Errorf("LAN sync request failed")
	}
	if nil == envelope.Data {
		return nil
	}
	data, err := gulu.JSON.MarshalJSON(envelope.Data)
	if err != nil {
		return err
	}
	return gulu.JSON.UnmarshalJSON(data, result)
}

func (c *lanCloud) CreateRepo(name string) (err error) {
	var result map[string]interface{}
	return c.call("/api/sync/lan/createRepo", map[string]interface{}{"name": name}, &result)
}

func (c *lanCloud) RemoveRepo(name string) (err error) {
	var result map[string]interface{}
	return c.call("/api/sync/lan/removeRepo", map[string]interface{}{"name": name}, &result)
}

func (c *lanCloud) GetRepos() (repos []*cloud.Repo, size int64, err error) {
	var result struct {
		Repos []*cloud.Repo `json:"repos"`
		Size  int64         `json:"size"`
	}
	err = c.call("/api/sync/lan/getRepos", map[string]interface{}{}, &result)
	return result.Repos, result.Size, err
}

func (c *lanCloud) UploadObject(filePath string, overwrite bool) (length int64, err error) {
	absFilePath := filepath.Join(c.Conf.RepoPath, filePath)
	data, err := os.ReadFile(absFilePath)
	if err != nil {
		return 0, err
	}
	return c.UploadBytes(filePath, data, overwrite)
}

func (c *lanCloud) UploadBytes(filePath string, data []byte, overwrite bool) (length int64, err error) {
	var result struct {
		Length int64 `json:"length"`
	}
	err = c.call("/api/sync/lan/uploadBytes", map[string]interface{}{
		"filePath":  filePath,
		"overwrite": overwrite,
		"data":      base64.StdEncoding.EncodeToString(data),
	}, &result)
	return result.Length, err
}

func (c *lanCloud) DownloadObject(filePath string) (data []byte, err error) {
	var result struct {
		Data string `json:"data"`
	}
	err = c.call("/api/sync/lan/downloadObject", map[string]interface{}{"filePath": filePath}, &result)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(result.Data)
}

func (c *lanCloud) RemoveObject(filePath string) (err error) {
	var result map[string]interface{}
	return c.call("/api/sync/lan/removeObject", map[string]interface{}{"filePath": filePath}, &result)
}

func (c *lanCloud) GetTags() (tags []*cloud.Ref, err error) {
	var result struct {
		Tags []*cloud.Ref `json:"tags"`
	}
	err = c.call("/api/sync/lan/getTags", map[string]interface{}{}, &result)
	return result.Tags, err
}

func (c *lanCloud) GetIndexes(page int) (indexes []*entity.Index, pageCount, totalCount int, err error) {
	var result struct {
		Indexes    []*entity.Index `json:"indexes"`
		PageCount  int             `json:"pageCount"`
		TotalCount int             `json:"totalCount"`
	}
	err = c.call("/api/sync/lan/getIndexes", map[string]interface{}{"page": page}, &result)
	return result.Indexes, result.PageCount, result.TotalCount, err
}

func (c *lanCloud) GetRefsFiles() (fileIDs []string, refs []*cloud.Ref, err error) {
	var result struct {
		FileIDs []string     `json:"fileIDs"`
		Refs    []*cloud.Ref `json:"refs"`
	}
	err = c.call("/api/sync/lan/getRefsFiles", map[string]interface{}{}, &result)
	return result.FileIDs, result.Refs, err
}

func (c *lanCloud) GetChunks(checkChunkIDs []string) (chunkIDs []string, err error) {
	var result struct {
		ChunkIDs []string `json:"chunkIDs"`
	}
	err = c.call("/api/sync/lan/getChunks", map[string]interface{}{"checkChunkIDs": checkChunkIDs}, &result)
	return result.ChunkIDs, err
}

func (c *lanCloud) GetStat() (stat *cloud.Stat, err error) {
	var result struct {
		Stat *cloud.Stat `json:"stat"`
	}
	err = c.call("/api/sync/lan/getStat", map[string]interface{}{}, &result)
	return result.Stat, err
}

func (c *lanCloud) GetAvailableSize() int64 {
	return c.Conf.AvailableSize
}

func (c *lanCloud) AddTraffic(traffic *cloud.Traffic) {}

func (c *lanCloud) ListObjects(pathPrefix string) (objInfos map[string]*entity.ObjectInfo, err error) {
	var result struct {
		Objects map[string]*entity.ObjectInfo `json:"objects"`
	}
	err = c.call("/api/sync/lan/listObjects", map[string]interface{}{"pathPrefix": pathPrefix}, &result)
	return result.Objects, err
}

func (c *lanCloud) GetIndex(id string) (index *entity.Index, err error) {
	var result struct {
		Index *entity.Index `json:"index"`
	}
	err = c.call("/api/sync/lan/getIndex", map[string]interface{}{"id": id}, &result)
	return result.Index, err
}

func (c *lanCloud) GetConcurrentReqs() int {
	return Conf.Sync.LAN.ConcurrentReqs
}

func BuildLANHostCloud(repoPath string) (*cloud.Local, error) {
	ensureLANConf()
	if !LANHostEnabled() {
		return nil, errors.New("LAN sync host is not enabled")
	}
	confLocal := &cloud.Conf{
		Dir:      Conf.Sync.CloudName,
		RepoPath: repoPath,
		Local: &cloud.ConfLocal{
			Endpoint:       Conf.Sync.LAN.BasePath,
			Timeout:        Conf.Sync.LAN.Timeout,
			ConcurrentReqs: Conf.Sync.LAN.ConcurrentReqs,
		},
	}
	return cloud.NewLocal(&cloud.BaseCloud{Conf: confLocal}), nil
}

func LanAuthOK(r *http.Request) bool {
	ensureLANConf()
	token := strings.TrimSpace(r.Header.Get("X-Padhivu-LAN-Token"))
	if "" == Conf.Sync.LAN.AuthToken {
		return "" == token || token == Conf.AccessAuthCode
	}
	return token == Conf.Sync.LAN.AuthToken
}

func TrackLANClient(r *http.Request) {
	registerLANDevice(&LANConnectedDevice{
		ID:        r.Header.Get("X-Padhivu-LAN-Device-ID"),
		Name:      r.Header.Get("X-Padhivu-LAN-Device-Name"),
		Host:      strings.Split(r.RemoteAddr, ":")[0],
		OS:        r.Header.Get("X-Padhivu-LAN-Device-OS"),
		Role:      "client",
		Connected: true,
	})
}

func LANPing(endpoint string, timeout int) error {
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Get(strings.TrimRight(endpoint, "/") + "/api/sync/lan/ping")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("LAN host check failed [%d] %s", resp.StatusCode, string(bytes.TrimSpace(body)))
	}
	return nil
}
