package main

import (
	_ "github.com/golangpkg/qor-cms/routers"
	"github.com/qor/admin"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/golangpkg/qor-cms/conf/auth"
	"github.com/golangpkg/qor-cms/conf/db"
	"github.com/golangpkg/qor-cms/models"
	"github.com/qor/auth/auth_identity"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

func main() {
	//开启session。配置文件 配置下sessionon = true即可。
	beego.BConfig.WebConfig.Session.SessionOn = true

	//DB, _ := gorm.Open("sqlite3", "demo.db") //for sqlite3
	DB := db.DB
	DB.AutoMigrate(&models.Category{}, &models.Article{}, &auth_identity.AuthIdentity{},
		&models.IndexSlider{}, &models.JobCompany{}, &models.Job{})

	// Initalize
	Admin := admin.New(&admin.AdminConfig{SiteName: "qor-cms", DB: DB, Auth: auth.AdminAuth{}})

	// Create resources from GORM-backend model
	//文章分类
	category := Admin.AddResource(&models.Category{}, &admin.Config{Name: "分类管理", Menu: []string{"资源管理"}})
	category.Meta(&admin.Meta{Name: "Name", Label: "名称"})
	//PageCount: 5,
	article := Admin.AddResource(&models.Article{}, &admin.Config{Name: "文章管理", Menu: []string{"资源管理"}})
	article.Meta(&admin.Meta{Name: "Title", Label: "标题", Type: "text"})
	article.Meta(&admin.Meta{Name: "ImgUrl", Label: "图片", Type: "kindimage"})
	article.Meta(&admin.Meta{Name: "Content", Label: "内容", Type: "kindeditor"})
	article.Meta(&admin.Meta{Name: "Category", Label: "分类"})
	article.Meta(&admin.Meta{Name: "CreatedAt", Label: "创建时间"})
	article.Meta(&admin.Meta{Name: "UpdatedAt", Label: "更新时间"})
	article.Meta(&admin.Meta{Name: "Url", Label: "地址", Type: "readonly"})
	article.Meta(&admin.Meta{Name: "IsPublish", Label: "是否发布", Type: "checkbox"})
	article.IndexAttrs("Title", "Category", "IsPublish", "Url", "ImgUrl", "CreatedAt", "UpdatedAt")
	//新增
	article.NewAttrs("Title", "Url", "IsPublish", "Category", "ImgUrl", "Content")
	//编辑
	article.EditAttrs("Title", "Url", "IsPublish", "Category", "ImgUrl", "Content")
	//增加发布功能：
	// 发布按钮，显示到右侧上面。
	article.Action(&admin.Action{
		Name:  "publishAll",
		Label: "全部发布",
		Handler: func(actionArgument *admin.ActionArgument) error {
			logs.Info("############### publishAll ###############")
			//生成html代码。
			models.GenArticleAndCategoryList(0)
			return nil
		},
		Modes: []string{"collection"},
	})
	// 发布按钮，显示到右侧上面。
	article.Action(&admin.Action{
		Name:  "publish5page",
		Label: "增量发布5页",
		Handler: func(actionArgument *admin.ActionArgument) error {
			logs.Info("############### publish5page ###############")
			//生成html代码。
			models.GenArticleAndCategoryList(5)
			return nil
		},
		Modes: []string{"collection"},
	})

	article.AddProcessor(&resource.Processor{
		Name: "process_store_data", // register another processor with
		Handler: func(value interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
			if article, ok := value.(*models.Article); ok {
				// do something...
				logs.Info("################ article ##################")
				if article.Url == "" {
					t := article.CreatedAt //time.Now()
					if t.IsZero() { //如果创建事件为空。
						t = time.Now()
					}
					url := fmt.Sprintf("%d-%02d/%d.html", t.Year(), t.Month(), t.Unix())
					logs.Info(t, url)
					article.Url = url
				}
				//更新摘要。新建，修改都更新。
				if article.Content != "" {
					//如果摘要为空，且内容不为空。
					//去除所有尖括号内的HTML代码，并换成换行符
					re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
					newContent := re.ReplaceAllString(article.Content, "\n")
					//去除连续的换行符
					re, _ = regexp.Compile("\\s{2,}")
					newContent = re.ReplaceAllString(newContent, "\n")
					newContent = strings.TrimSpace(newContent)
					newContentRune := []rune(newContent)
					if len(newContentRune) > 75 {
						article.Description = string(newContentRune[0:75])
					} else {
						article.Description = newContent
					}
					logs.Info("description: ", article.Description)
				}

			}
			return nil
		},
	})

	//增加首页轮播图。
	indexSlider := Admin.AddResource(&models.IndexSlider{}, &admin.Config{Name: "首页轮播图", Menu: []string{"资源管理"}})
	indexSlider.Meta(&admin.Meta{Name: "Image", Label: "图片（860X300）地址", Type: "kindimage"})
	indexSlider.Meta(&admin.Meta{Name: "Url", Label: "链接地址"})
	indexSlider.Action(&admin.Action{
		Name:  "publish",
		Label: "发布",
		Handler: func(actionArgument *admin.ActionArgument) error {
			logs.Info("############### publish ###############")
			//生成html代码。
			models.GenIndexSlider()
			return nil
		},
		Modes: []string{"collection"},
	})

	//################################ 招聘模块 ################################
	jobCompany := Admin.AddResource(&models.JobCompany{}, &admin.Config{Name: "招聘公司管理", Menu: []string{"招聘管理"}})
	jobCompany.Meta(&admin.Meta{Name: "Name", Label: "公司名称"})
	jobCompany.Meta(&admin.Meta{Name: "IndustryType", Label: "行业分类", Config: &admin.SelectOneConfig{
		Collection: []string{"计算机/互联网/通信/电子", "会计/金融/银行/保险", "贸易/消费/制造/营运", "制药/医疗",
			"广告/媒体", "房地产/建筑", "专业服务/教育/培训", "服务业", "物流/运输", "能源/原材料", "政府/非营利组织/其他"}}})
	jobCompany.Meta(&admin.Meta{Name: "CompanyInfo", Label: "公司描述", Type: "kindeditor"})

	job := Admin.AddResource(&models.Job{}, &admin.Config{Name: "招聘职位管理", Menu: []string{"招聘管理"}})
	job.Meta(&admin.Meta{Name: "IsPublish", Label: "是否发布", Type: "checkbox"})
	job.Meta(&admin.Meta{Name: "Title", Label: "标题", Type: "text"})
	job.Meta(&admin.Meta{Name: "Salary", Label: "薪水（/月）", Type: "text"})
	job.Meta(&admin.Meta{Name: "JobCompany", Label: "公司名称"})
	job.Meta(&admin.Meta{Name: "Locale", Label: "工作地点", Config: &admin.SelectOneConfig{
		Collection: []string{"北京", "上海", "广州", "深圳", "天津", "杭州", "成都"}}})
	job.Meta(&admin.Meta{Name: "Education", Label: "学历", Config: &admin.SelectOneConfig{
		Collection: []string{"大专", "本科", "硕士", "博士"}}})
	job.Meta(&admin.Meta{Name: "Age", Label: "年龄（岁）", Type: "text"})
	job.Meta(&admin.Meta{Name: "WorkExperience", Label: "工作年限", Config: &admin.SelectOneConfig{
		Collection: []string{"1-3年", "3-5年", "5-10年", "10年以上"}}})
	job.Meta(&admin.Meta{Name: "Department", Label: "所属部门", Type: "text"})
	job.Meta(&admin.Meta{Name: "ReportTo", Label: "汇报对象", Type: "text"})
	job.Meta(&admin.Meta{Name: "PublishDate", Label: "发部日期", Type: "date"})
	job.Meta(&admin.Meta{Name: "JobInfo", Label: "职位描述", Type: "kindeditor"})
	job.IndexAttrs("Title", "Salary", "JobCompany", "Locale", "Education", "Age", "WorkExperience")
	//新增
	job.NewAttrs("IsPublish", "Title", "Salary", "JobCompany", "Locale", "Education", "Age", "WorkExperience",
		"Department", "ReportTo", "PublishDate", "JobInfo")
	//编辑
	job.EditAttrs("IsPublish", "Title", "Salary", "JobCompany", "Locale", "Education", "Age", "WorkExperience",
		"Department", "ReportTo", "PublishDate", "JobInfo")
	//增加发布功能：
	// 发布按钮，显示到右侧上面。
	job.Action(&admin.Action{
		Name:  "publishJobList",
		Label: "发布",
		Handler: func(actionArgument *admin.ActionArgument) error {
			logs.Info("############### publishJobList ###############")
			//生成html代码。
			models.GenJobList()
			return nil
		},
		Modes: []string{"collection"},
	})

	// 启动服务
	mux := http.NewServeMux()
	Admin.MountTo("/admin", mux)
	beego.Handler("/admin/*", mux)
	beego.Run()
}
