package api

import (
	"encoding/base64"
	"net/http"

	"github.com/88250/gulu"
	"github.com/gin-gonic/gin"
	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/siyuan-note/siyuan/kernel/util"
)

func setSyncProviderLAN(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	lanArg := arg["lan"]
	data, err := gulu.JSON.MarshalJSON(lanArg)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	lan := &conf.LAN{}
	if err = gulu.JSON.UnmarshalJSON(data, lan); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	if err = model.SetSyncProviderLAN(lan); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		ret.Data = map[string]interface{}{"closeTimeout": 5000}
		return
	}
	ret.Data = map[string]interface{}{"lan": lan}
}

func lanPing(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"ok":       true,
		"kernel":   model.KernelID,
		"device":   model.Conf.System.Name,
		"host":     model.LANHostEnabled(),
		"provider": model.Conf.Sync.Provider,
	})
}

func lanCreateRepo(c *gin.Context)     { lanLocalCall(c, "createRepo") }
func lanRemoveRepo(c *gin.Context)     { lanLocalCall(c, "removeRepo") }
func lanGetRepos(c *gin.Context)       { lanLocalCall(c, "getRepos") }
func lanUploadBytes(c *gin.Context)    { lanLocalCall(c, "uploadBytes") }
func lanDownloadObject(c *gin.Context) { lanLocalCall(c, "downloadObject") }
func lanRemoveObject(c *gin.Context)   { lanLocalCall(c, "removeObject") }
func lanGetTags(c *gin.Context)        { lanLocalCall(c, "getTags") }
func lanGetIndexes(c *gin.Context)     { lanLocalCall(c, "getIndexes") }
func lanGetRefsFiles(c *gin.Context)   { lanLocalCall(c, "getRefsFiles") }
func lanGetChunks(c *gin.Context)      { lanLocalCall(c, "getChunks") }
func lanGetStat(c *gin.Context)        { lanLocalCall(c, "getStat") }
func lanListObjects(c *gin.Context)    { lanLocalCall(c, "listObjects") }
func lanGetIndex(c *gin.Context)       { lanLocalCall(c, "getIndex") }

func lanLocalCall(c *gin.Context, method string) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	if !model.LANHostEnabled() || !model.LanAuthOK(c.Request) {
		ret.Code = -1
		ret.Msg = "LAN sync host unavailable or unauthorized"
		return
	}
	model.TrackLANClient(c.Request)

	arg, _ := util.JsonArg(c, ret)
	local, err := model.BuildLANHostCloud(util.RepoDir)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	switch method {
	case "createRepo":
		ret.Code = errorCode(local.CreateRepo(arg["name"].(string)), ret)
	case "removeRepo":
		ret.Code = errorCode(local.RemoveRepo(arg["name"].(string)), ret)
	case "getRepos":
		repos, size, callErr := local.GetRepos()
		ret.Code = errorCode(callErr, ret)
		ret.Data = map[string]interface{}{"repos": repos, "size": size}
	case "uploadBytes":
		data, decodeErr := base64.StdEncoding.DecodeString(arg["data"].(string))
		if decodeErr != nil {
			ret.Code = -1
			ret.Msg = decodeErr.Error()
			return
		}
		length, callErr := local.UploadBytes(arg["filePath"].(string), data, arg["overwrite"].(bool))
		ret.Code = errorCode(callErr, ret)
		ret.Data = map[string]interface{}{"length": length}
	case "downloadObject":
		data, callErr := local.DownloadObject(arg["filePath"].(string))
		ret.Code = errorCode(callErr, ret)
		ret.Data = map[string]interface{}{"data": base64.StdEncoding.EncodeToString(data)}
	case "removeObject":
		ret.Code = errorCode(local.RemoveObject(arg["filePath"].(string)), ret)
	case "getTags":
		tags, callErr := local.GetTags()
		ret.Code = errorCode(callErr, ret)
		ret.Data = map[string]interface{}{"tags": tags}
	case "getIndexes":
		indexes, pageCount, totalCount, callErr := local.GetIndexes(int(arg["page"].(float64)))
		ret.Code = errorCode(callErr, ret)
		ret.Data = map[string]interface{}{"indexes": indexes, "pageCount": pageCount, "totalCount": totalCount}
	case "getRefsFiles":
		fileIDs, refs, callErr := local.GetRefsFiles()
		ret.Code = errorCode(callErr, ret)
		ret.Data = map[string]interface{}{"fileIDs": fileIDs, "refs": refs}
	case "getChunks":
		chunkInterfaces := arg["checkChunkIDs"].([]interface{})
		checkChunkIDs := make([]string, 0, len(chunkInterfaces))
		for _, chunkID := range chunkInterfaces {
			checkChunkIDs = append(checkChunkIDs, chunkID.(string))
		}
		chunkIDs, callErr := local.GetChunks(checkChunkIDs)
		ret.Code = errorCode(callErr, ret)
		ret.Data = map[string]interface{}{"chunkIDs": chunkIDs}
	case "getStat":
		stat, callErr := local.GetStat()
		ret.Code = errorCode(callErr, ret)
		ret.Data = map[string]interface{}{"stat": stat}
	case "listObjects":
		objects, callErr := local.ListObjects(arg["pathPrefix"].(string))
		ret.Code = errorCode(callErr, ret)
		ret.Data = map[string]interface{}{"objects": objects}
	case "getIndex":
		index, callErr := local.GetIndex(arg["id"].(string))
		ret.Code = errorCode(callErr, ret)
		ret.Data = map[string]interface{}{"index": index}
	}
}

func errorCode(err error, ret *gulu.Result) int {
	if err != nil {
		ret.Msg = err.Error()
		return -1
	}
	return 0
}
