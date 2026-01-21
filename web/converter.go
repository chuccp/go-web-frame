package web

//type Converter func(value any, err error, ctx *HttpContext)
//
//var DefaultConverter Converter = func(value any, err error, ctx_ *HttpContext) {
//	response := ctx_.Response()
//	if err != nil {
//		err0 := Errors(value, err)
//		response.JSON(err0.Code, err0)
//		response.Abort()
//	} else {
//		if value != nil {
//			switch t := value.(type) {
//			case *Message:
//				if t.Code == http.StatusMovedPermanently {
//					response.Redirect(http.StatusMovedPermanently, t.Data.(string))
//					response.Abort()
//					return
//				}
//				response.JSON(t.Code, value)
//			case string:
//				_, err2 := response.Write([]byte(t))
//				if err2 != nil {
//					response.Abort()
//					return
//				}
//			case *File:
//				if len(t.FileName) == 0 {
//					_, filename := path.Split(t.Path)
//					t.FileName = filename
//				}
//				if util.IsNotBlank(t.Suffix) && !strings.HasSuffix(t.FileName, t.Suffix) {
//					if !strings.HasPrefix(t.Suffix, ".") {
//						t.Suffix = "." + t.Suffix
//					}
//					t.FileName = t.FileName + t.Suffix
//				}
//				response.FileAttachment(t.Path, t.FileName)
//			case *os.File:
//				response.FileAttachment(t.Name(), t.Name())
//			default:
//				response.JSON(200, Data(value))
//			}
//		}
//	}
//}
