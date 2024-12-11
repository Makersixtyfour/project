package app

import (
	"context"
	"dannyroman2015/phoebe/internal/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
	IsAdmin  string `bson:"isadmin"`
	Info     struct {
		StaffId string `bson:"staffid"`
		Name    string `bson:"name"`
	} `bson:"info"`
	Defaulturl string   `bson:"defaulturl"`
	Authurls   []string `bson:"authurls"`
}

type CuttingRecord struct {
	Type         string    `bson:"type"`
	Date         time.Time `bson:"date"`
	Qty          float64   `bson:"qty"`
	Unit         string    `bson:"unit"`
	Reporter     string    `bson:"reporter"`
	CreatedDate  time.Time `bson:"createdDate"`
	ModifiedDate time.Time `bson:"modifiedDate"`
}

type PackingRecord struct {
	Date     time.Time `bson:"date"`
	ProType  string    `bson:"protype"`
	FacNo    string    `bson:"facno"`
	Qty      int       `bson:"qty"`
	Unit     string    `bson:"unit"`
	Price    float64   `bson:"price"`
	Currency string    `bson:"currency"`
}

type Score struct {
	EmpId      string `bson:"empid"`
	EmpName    string `bson:"empname"`
	EmpSection string `bson:"empsection"`
	PointTotal int    `bson:"point_total"`
}

type Employee struct {
	Id      string `bson:"id"`
	Name    string `bson:"name"`
	Section string `bson:"section"`
}

type CuttingQty struct {
	Date time.Time `bson:"date"`
	Qty  float64   `bson:"qty_total"`
}
type Server struct {
	Addr   string
	Logger *log.Logger
	pgdb   *sql.DB
	mgdb   *mongo.Database
}

func NewServer(addr string, mgdb *mongo.Database) *Server {
	logger := log.New(os.Stdout, "Logger: ", log.LstdFlags)

	return &Server{
		Addr:   addr,
		Logger: logger,
		mgdb:   mgdb,
	}
}

func (s *Server) Start() {
	log.Printf("Starting server on %s", s.Addr)

	srv := &http.Server{
		Addr:         s.Addr,
		Handler:      s.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	err := srv.ListenAndServe()
	log.Fatal(err)
}

func (s *Server) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(s.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(s.methodNotAllowedResponse)

	router.ServeFiles("/static/*filepath", http.Dir("static"))

	router.GET("/", s.index)
	router.GET("/home", s.home)

	router.GET("/login", s.serveLogin)
	router.POST("/login", s.requestLogin)
	router.GET("/logout", s.logout)
	router.POST("/logout", s.logout)
	// router.POST("/request", s.sendRequest)

	// router.GET("/admin", withAuth(s.admin))

	router.GET("/dashboard", s.dashboard)
	// router.GET("/dashboard/loadotattend", s.d_loadotattend)
	// router.GET("/dashboard/loadmanpower", s.d_loadmanpower)
	router.GET("/dashboard/loadproductionvop", s.d_loadproductionvop)
	// router.GET("/dashboard/loadplan", s.d_loadproductionplan)
	router.GET("/dashboard/loadrawwood", s.d_loadrawwood)
	router.GET("/dashboard/loadproduction", s.d_loadproduction)
	router.GET("/dashboard/loadreededline", s.d_loadreededline)
	router.GET("/dashboard/loadoutput", s.d_loadoutput)
	router.GET("/dashboard/loadpanelcnc", s.d_loadpanelcnc)
	router.GET("/dashboard/loadveneer", s.d_loadveneer)
	router.GET("/dashboard/loadassembly", s.d_loadassembly)
	router.GET("/dashboard/loadwhitewood", s.d_loadwhitewood)
	router.GET("/dashboard/loadwoodfinish", s.d_loadwoodfinish)
	// router.GET("/dashboard/loadfinemill", s.d_loadfinemill)
	router.GET("/dashboard/loadpack", s.d_loadpack)
	router.GET("/dashboard/loadslicing", s.d_loadslicing)
	router.GET("/dashboard/loadquality", s.d_loadquality)
	router.GET("/dashboard/loadwoodrecovery", s.d_loadwoodrecovery)
	// router.GET("/dashboard/loaddowntime", s.d_loaddowntime)
	// router.GET("/dashboard/loadsixs", s.d_loadsixs)
	// router.GET("/dashboard/loadsafety", s.d_loadsafety)
	// router.POST("/dashboard/otattend/getchart", s.doa_getchart)
	router.POST("/dashboard/manpower/getchart", s.dmp_getchart)
	router.POST("/dashboard/productionvop/getchart", s.dpv_getchart)
	router.POST("/dashboard/production/getchart", s.dpr_getchart)
	router.POST("/dashboard/panelcnc/getchart", s.dpc_getchart)
	router.POST("/dashboard/assembly/getchart", s.da_getchart)
	router.POST("/dashboard/woodfinish/getchart", s.dw_getchart)
	router.POST("/dashboard/whitewood/getchart", s.dww_getchart)
	router.POST("/dashboard/cutting/getchart", s.dc_getchart)
	router.POST("/dashboard/lamination/getchart", s.dl_getchart)
	router.POST("/dashboard/reededline/getchart", s.dr_getchart)
	router.POST("/dashboard/output/getchart", s.do_getchart)
	router.POST("/dashboard/veneer/getchart", s.dv_getchart)
	// router.POST("/dashboard/finemill/getchart", s.df_getchart)
	router.POST("/dashboard/pack/getchart", s.dp_getchart)
	router.POST("/dashboard/woodrecovery/getchart", s.dwr_getchart)
	// router.POST("/dashboard/sixs/getchart", s.ds_getchart)
	router.POST("/dashboard/quality/getchart", s.dq_getchart)
	// router.POST("/dashboard/downtime/getchart", s.dd_getchart)
	// router.POST("/dashboard/safety/getchart", s.dst_getchart)

	router.GET("/incentive/admin", withAuth(s.iadmin))
	router.GET("/incentive/admin/loadcrittable", s.loadcrittable)
	router.POST("/incentive/admin/upsertcriteria", s.caupsertcriteria)
	router.DELETE("/incentive/admin/deletecriteria/:criteriaid", s.deletecriteria)
	router.POST("/incentive/admin/searchcriterion", s.ia_searchcriterion)
	router.GET("/incentive/admin/loadevaltable", s.loadevaltable)
	router.DELETE("/incentive/admin/deleteevaluate/:evaluateid", s.deleteevaluate)
	router.POST("/incentive/admin/searchevaluate", s.ia_searchevaluate)

	router.GET("/incentive/evaluate", s.evaluate) // use withAuth later
	router.POST("/incentive/evaluate/searchstaff", s.searchstaff)
	router.POST("/incentive/evaluate/searchcriterion", s.searchcriterion)
	router.POST("/incentive/evaluate/sendevaluate", s.sendevaluate)

	router.GET("/incentive/overview", s.ioverview)
	router.GET("/incentive/overview/loadscores", s.io_loadscores)
	router.POST("/incentive/overview/scoresearch", s.io_scoresearch)
	router.POST("/incentive/overview/evalsearch", s.io_evalsearch)

	router.GET("/hr/overview", s.hr_overview)
	router.GET("/hr/overview/loadchart", s.hr_loadchart)

	router.GET("/hr/admin", withAuth(s.hradmin))
	router.POST("/hr/admin/searchemployee", s.ha_searchemployee)
	router.POST("/hr/admin/upsertemployee", s.ha_upsertemployee)
	router.GET("/hr/admin/exportempexcel", s.ha_exportempexcel)
	router.GET("/hr/admin/prevnext/:currentPage/:prevnext", s.ha_prevnext)
	router.GET("/hr/entry", withAuth(s.hr_entry))
	router.POST("/hr/entry", s.hr_insertemplist)
	router.GET("/6s/overview", s.s_overview)
	router.GET("/6s/entry", s.s6_entry)
	router.POST("/6s/entry", s.s6_sendentry)
	router.GET("/6s/admin", s.s6_admin)

	// Rawwood ////////////////////////////////////////////////////////
	router.POST("/sections/rawwood/entry/entry", withAuth2(s.sre_entry))
	router.POST("/sections/rawwood/entry/selection", withAuth2(s.sre_selection))
	// End Rawwood ////////////////////////////////////////////////////////

	// Cuttting ////////////////////////////////////////////////////////
	router.GET("/sections/cutting/overview", s.sc_overview)
	router.GET("/sections/cutting/overview/loadwoodremain", s.sco_loadwoodremain)
	router.GET("/sections/cutting/overview/loadwrnote", s.sco_loadwrnote)
	router.GET("/sections/cutting/overview/loadreport", s.sco_loadreport)
	router.POST("/sections/cutting/overview/wrnotesearch", s.sco_wrnotesearch)
	router.POST("/sections/cutting/overview/reportsearch", s.sco_reportsearch)
	router.POST("/sections/cutting/overview/wrnotefilter", s.sco_wrnotefilter)
	router.POST("/sections/cutting/overview/reportfilter", s.sco_reportfilter)
	router.POST("/sections/cutting/overview/createdemand", s.sco_createdemand)
	router.PUT("/sections/cutting/overview/wrnotereturn/:wrnotecode/:remainqty", s.sco_wrnotereturn) // cho huế

	router.GET("/sections/cutting/entry", withAuth(s.sc_entry))
	router.POST("/sections/cutting/entry/wrnoteinfo", s.sc_wrnoteinfo)
	router.GET("/sections/cutting/entry/newwrnote", s.sc_newwrnote)
	router.POST("/sections/cutting/entry/createwrnote", s.sc_createwrnote)
	router.POST("/sections/cutting/sendentry", s.sc_sendentry)
	router.GET("/sections/cutting/entry/return", s.sce_return)
	router.POST("/sections/cutting/entry/sendreturn", s.sce_sendreturn)
	router.GET("/sections/cutting/entry/fine", s.sce_fine)
	router.POST("/sections/cutting/entry/sendfine", s.sce_sendfine)

	router.GET("/sections/cutting/woodrecoveryentry", withAuth(s.sc_woodrecoveryentry))
	router.GET("/sections/cutting/entry/wr_loadform", s.sce_wr_loadform)
	router.POST("/sections/cutting/entry/wr_sendentry", s.sce_wr_sendentry)

	router.GET("/sections/cutting/admin", withAuth(s.sc_admin))
	router.GET("/sections/cutting/admin/loadreports", s.sc_loadreports)
	router.GET("/sections/cutting/admin/loadwrnote", s.sc_loadwrnote)
	router.POST("/sections/cutting/admin/searchreport", s.sca_searchreport)
	router.POST("/sections/cutting/admin/searchwrnote", s.sca_searchwrnote)
	router.POST("/sections/cutting/admin/reportdatefilter", s.sca_reportdatefilter)
	router.DELETE("/sections/cutting/admin/deletereport/:reportid", s.sca_deletereport)
	router.DELETE("/sections/cutting/admin/deletewrnote/:wrnoteid", s.sca_deletewrnote)
	router.GET("/sections/cutting/admin/wrnoteupdateform/:wrnoteid", s.sca_wrnoteupdateform)
	router.PUT("/sections/cutting/admin/updatewrnote/:wrnoteid", s.sca_updatewrnote)
	// end Cuttting/////////////////////////////////////////////////////////////

	// Lamination ////////////////////////////////////////////////////////
	router.GET("/sections/lamination/overview", s.sl_overview)
	router.GET("/sections/lamination/overview/loadreport", s.slo_loadreport)
	router.POST("/sections/lamination/overview/reportsearch", s.slo_reportsearch)
	router.POST("/sections/lamination/overview/reportdatefilter", s.slo_reportdatefilter)

	router.GET("/sections/lamination/entry", withAuth(s.sl_entry))
	router.GET("/sections/lamination/entry/loadform", s.sle_loadform)
	router.POST("/sections/lamination/entry/sendentry", s.sle_sendentry)

	router.GET("/sections/lamination/admin", withAuth(s.sl_admin))
	router.GET("/sections/lamination/admin/loadreport", s.sla_loadreport)
	router.POST("/sections/lamination/admin/searchreport", s.sla_searchreport)
	router.DELETE("/sections/lamination/admin/deletereport/:reportid", s.sla_deletereport)
	// end Lamination/////////////////////////////////////////////////////////////

	// Slicing ////////////////////////////////////////////////////////
	router.GET("/sections/slicing/entry", withAuth(s.ss_entry))
	router.GET("/sections/slicing/entry/loadform", s.sse_loadform)
	router.POST("/sections/slicing/entry/sendentry", s.sse_sendentry)

	router.GET("/sections/slicing/admin", s.ss_admin)
	router.GET("/sections/slicing/admin/loadreport", s.ssa_loadreport)
	router.POST("/sections/slicing/admin/reportsearch", s.ssa_reportsearch)
	router.DELETE("/sections/slicing/admin/deletereport/:id", s.ssa_deletereport)
	// end Slicing/////////////////////////////////////////////////////////////

	// Reededline ////////////////////////////////////////////////////////
	router.GET("/sections/reededline/overview", s.sr_overview)
	router.GET("/sections/reededline/overview/loadreport", s.sro_loadreport)
	router.POST("/sections/reededline/overview/reportsearch", s.sro_reportsearch)
	router.POST("/sections/reededline/overview/reportdatefilter", s.sro_reportdatefilter)

	router.GET("/sections/reededline/entry", withAuth(s.sr_entry))
	router.GET("/sections/reededline/entry/loadform", s.sre_loadform)
	router.POST("/sections/reededline/entry/sendentry", s.sre_sendentry)

	router.GET("/sections/reededline/admin", withAuth(s.sr_admin))
	router.GET("/sections/reededline/admin/loadreport", s.sra_loadreport)
	router.POST("/sections/reededline/admin/searchreport", s.sra_searchreport)
	router.DELETE("/sections/reededline/admin/deletereport/:reportid", s.sra_deletereport)
	// end Reededline/////////////////////////////////////////////////////////////

	// Output ////////////////////////////////////////////////////////
	router.GET("/sections/output/entry", withAuth(s.so_entry))
	router.GET("/sections/output/entry/loadentry", s.soe_loadentry)
	router.POST("/sections/output/entry/sendentry", s.soe_sendentry)
	router.GET("/sections/output/entry/loadformentry", s.soe_loadformentry)
	router.POST("/sections/output/entry/sendfastentry", s.soe_sendfastentry)

	router.GET("/sections/output/admin", withAuth(s.so_admin))
	router.GET("/sections/output/admin/loadreport", s.soa_loadreport)
	router.POST("/sections/output/admin/searchreport", s.soa_searchreport)
	router.POST("/sections/output/admin/reportdatefilter", s.soa_reportdatefilter)
	router.DELETE("/sections/output/admin/deletereport/:reportid", s.soa_deletereport)
	router.GET("/sections/output/admin/updateform/:reportid", s.soa_updateform)
	router.PUT("/sections/output/admin/updatereport/:reportid", s.soa_updatereport)

	// end Output/////////////////////////////////////////////////////////////

	// Veneer ////////////////////////////////////////////////////////
	router.GET("/sections/veneer/overview", s.sv_overview)
	router.GET("/sections/veneer/overview/loadreport", s.svo_loadreport)
	router.POST("/sections/veneer/overview/reportsearch", s.svo_reportsearch)
	router.POST("/sections/veneer/overview/reportdatefilter", s.svo_reportdatefilter)

	router.GET("/sections/veneer/entry", withAuth(s.sv_entry))
	router.GET("/sections/veneer/entry/loadform", s.sve_loadform)
	router.POST("/sections/veneer/entry/sendentry", s.sve_sendentry)

	router.GET("/sections/veneer/admin", withAuth(s.sv_admin))
	router.GET("/sections/veneer/admin/loadreport", s.sva_loadreport)
	router.POST("/sections/veneer/admin/searchreport", s.sva_searchreport)
	router.DELETE("/sections/veneer/admin/deletereport/:reportid", s.sva_deletereport)
	// end Veneer/////////////////////////////////////////////////////////////

	router.GET("/sections/finemill/entry", withAuth(s.sf_entry))
	router.GET("/sections/finemill/entry/loadform", s.sfe_loadform)
	router.POST("/sections/finemill/entry/sendentry", s.sfe_sendentry)

	// Assembly ////////////////////////////////////////////////////////
	router.GET("/sections/assembly/overview", s.sa_overview)
	router.GET("/sections/assembly/overview/loadreport", s.sao_loadreport)
	router.POST("/sections/assembly/overview/reportsearch", s.sao_reportsearch)
	router.POST("/sections/assembly/overview/reportdatefilter", s.sao_reportdatefilter)
	router.POST("/sections/assembly/overview/addplanvalue", s.sao_addplanvalue)
	router.POST("/sections/assembly/overview/updateinventory", s.sao_updateinventory)

	router.GET("/sections/assembly/entry", withAuth(s.sa_entry))
	router.GET("/sections/assembly/entry/loadform", s.sae_loadform)
	router.POST("/sections/assembly/entry/sendentry", s.sae_sendentry)

	router.GET("/sections/assembly/planentry", s.sae_planentry)
	router.GET("/sections/assembly/entry/loadplanform", s.sae_loadplanform)
	router.POST("/sections/assembly/entry/sendplanentry", s.sae_sendplanentry)

	router.GET("/sections/assembly/inventoryentry", s.sai_inventoryentry)
	router.GET("/sections/assembly/entry/loadinventoryform", s.sai_loadinventoryform)

	router.GET("/sections/assembly/admin", withAuth(s.sa_admin))
	router.GET("/sections/assembly/admin/loadreport", s.saa_loadreport)
	router.POST("/sections/assembly/admin/searchreport", s.saa_searchreport)
	router.DELETE("/sections/assembly/admin/deletereport/:reportid", s.saa_deletereport)
	// end Assembly/////////////////////////////////////////////////////////////

	// WoodFinish ////////////////////////////////////////////////////////
	router.GET("/sections/woodfinish/overview", s.sw_overview)
	router.GET("/sections/woodfinish/overview/loadreport", s.swo_loadreport)
	router.POST("/sections/woodfinish/overview/reportsearch", s.swo_reportsearch)
	router.POST("/sections/woodfinish/overview/reportdatefilter", s.swo_reportdatefilter)

	router.GET("/sections/woodfinish/entry", withAuth(s.sw_entry))
	router.GET("/sections/woodfinish/entry/loadform", s.swe_loadform)
	router.POST("/sections/woodfinish/entry/sendentry", s.swe_sendentry)

	router.POST("/sections/woodfinish/overview/addplanvalue", s.swo_addplanvalue)
	router.POST("/sections/woodfinish/overview/updateinventory", s.swo_updateinventory)

	router.GET("/sections/woodfinish/admin", withAuth(s.sw_admin))
	router.GET("/sections/woodfinish/admin/loadreport", s.swa_loadreport)
	router.POST("/sections/woodfinish/admin/searchreport", s.swa_searchreport)
	router.DELETE("/sections/woodfinish/admin/deletereport/:reportid", s.swa_deletereport)
	// end WoodFinish/////////////////////////////////////////////////////////////

	// White Wood ////////////////////////////////////////////////////////
	router.POST("/sections/whitewood/overview/addmoney", withAuth2(s.swo_addmoney))
	router.POST("/sections/whitewood/overview/addplanvalue", withAuth2(s.swwo_addplanvalue))
	router.POST("/sections/whitewood/overview/updateinventory", s.swwo_updateinventory)
	router.POST("/sections/whitewood/overview/addnammoney", s.swo_addnammoney)
	// end White Wood/////////////////////////////////////////////////////////////

	// Pack ////////////////////////////////////////////////////////
	router.GET("/sections/pack/overview", s.spk_overview)
	router.GET("/sections/pack/overview/loadreport", s.pko_loadreport)
	router.POST("/sections/pack/overview/reportsearch", s.pko_reportsearch)
	router.POST("/sections/pack/overview/reportdatefilter", s.pko_reportdatefilter)

	router.GET("/sections/pack/entry", withAuth(s.spk_entry))
	router.GET("/sections/pack/entry/loadform", s.spk_loadform)
	router.POST("/sections/pack/entry/sendentry", s.spk_sendentry)

	router.POST("/sections/pack/overview/updateinventory", s.spo_updateinventory)
	router.POST("/sections/pack/overview/addplanvalue", s.spo_addplanvalue)

	router.GET("/sections/pack/admin", withAuth(s.spk_admin))
	router.GET("/sections/pack/admin/loadreport", s.spka_loadreport)
	router.POST("/sections/pack/admin/searchreport", s.spka_searchreport)
	router.POST("/sections/pack/admin/reportdatefilter", s.spka_reportdatefilter)
	router.DELETE("/sections/pack/admin/deletereport/:reportid", s.spka_deletereport)
	router.GET("/sections/pack/admin/updateform/:reportid", s.spka_updateform)
	router.PUT("/sections/pack/admin/updatereport/:reportid", s.spka_updatereport)
	// end Pack/////////////////////////////////////////////////////////////

	// Panelcnc ////////////////////////////////////////////////////////
	router.GET("/sections/panelcnc/overview", s.spc_overview)
	router.GET("/sections/panelcnc/overview/loadreport", s.spco_loadreport)
	router.POST("/sections/panelcnc/overview/reportsearch", s.spco_reportsearch)
	router.POST("/sections/panelcnc/overview/reportdatefilter", s.spco_reportdatefilter)

	router.GET("/sections/panelcnc/entry", withAuth(s.spc_entry))
	router.GET("/sections/panelcnc/entry/loadform", s.spc_loadform)
	router.POST("/sections/panelcnc/entry/sendentry", s.spc_sendentry)

	router.GET("/sections/panelcnc/admin", withAuth(s.spc_admin))
	router.GET("/sections/panelcnc/admin/loadreport", s.spca_loadreport)
	router.POST("/sections/panelcnc/admin/searchreport", s.spca_searchreport)
	router.DELETE("/sections/panelcnc/admin/deletereport/:reportid", s.spca_deletereport)
	// end Panelcnc/////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////////
	// Outsource ////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////
	router.GET("/sections/outsource/entry", withAuth(s.sos_entry))
	router.GET("/sections/outsource/entry/loadform", s.sose_loadform)
	router.POST("/sections/outsource/entry/sendentry", s.sose_sendentry)
	// end Outsource

	router.GET("/sections/packing/overview", s.sp_overview)

	router.GET("/sections/packing/entry", withAuth(s.sp_entry))
	router.GET("/sections/packing/entry/loadentry", s.sp_loadentry)
	router.GET("/selections/packing/entry/mo/:status", s.sp_mobystatus)
	router.POST("/selections/packing/entry/mosearch", s.sp_mosearch)
	router.GET("/sections/packing/entry/itemparts/:mo/:itemid/:pi", s.sp_itemparts)
	router.POST("/sections/packing/entry/initpart", s.sp_itempart)
	router.POST("/sections/packing/entry/initparts", s.sp_initparts)
	router.POST("/sections/packing/entry/maxpartqtyinput", s.sp_getinputmax)
	router.POST("/sections/packing/sendentry", s.sp_sendentry)

	router.GET("/sections/packing/admin", s.sp_admin)

	////////////////////////////////////////////////////////////////////
	// Production Value ////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////
	router.GET("/production/overview", s.p_overview)
	router.GET("/production/overview/loadprodtype", s.po_loadprodtype)
	router.GET("/production/overview/loadsummary", s.po_loadsummary)
	router.GET("/production/overview/loadreport", s.po_loadreport)
	router.POST("/production/overview/reportfilter", s.po_reportfilter)
	// router.POST("/production/overview/reportsearch", s.po_reportsearch)
	router.POST("/production/overview/prodtypefilter", s.po_prodtypefilter)
	router.POST("/production/overview/summarydatefilter", s.po_summarydatefilter)

	// end Production Value

	////////////////////////////////////////////////////////////////////
	// Target ////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////
	router.GET("/target/entry", withAuth(s.tg_entry))
	router.GET("/target/entry/loadsectionentry", s.tge_loadsectionentry)
	router.GET("/target/entry/loadreport", s.tge_loadreport)
	router.POST("/target/entry/settarget", s.tge_settarget)
	router.GET("/target/entry/loadplanworkdays", s.tge_loadplanworkdays)
	router.POST("/target/entry/setworkdays", s.tge_setworkdays)
	router.POST("/target/entry/search", s.tge_search)
	router.POST("/target/entry/filterbydate", s.tge_filterbydate)
	router.DELETE("/target/entry/deletereport/:id", s.tge_deletereport)

	// end Target

	////////////////////////////////////////////////////////////////////
	// Manhr ////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////
	router.GET("/manhr/admin", withAuth(s.m_admin))
	router.GET("/manhr/admin/loadentry", s.ma_loadentry)
	router.GET("/manhr/admin/loadreport", s.ma_loadreport)
	router.POST("/manhr/admin/sendentry", s.ma_sendentry)
	router.DELETE("/manhr/admin/deletereport/:id", s.ma_deletereport)
	router.GET("/manhr/admin/updateform/:id", s.ma_updateform)
	router.PUT("/manhr/admin/updatereport/:id", s.ma_updatereport)
	router.POST("/manhr/admin/search", s.ma_search)
	router.POST("/manhr/admin/filterbydate", s.ma_filterbydate)
	router.POST("/manhr/entry/sendtotalmanhr", withAuth2(s.me_sendtotalmanhr))
	// end Manhr

	router.GET("/downtime/entry", withAuth(s.dt_entry))
	router.GET("/downtime/entry/loadform", s.dte_loadform)
	router.POST("/downtime/entry/sendentry", s.dte_sendentry)

	router.POST("/mixingcolor/getusingstart", s.getusingstart)
	router.GET("/mixingcolor/loaddeliveryentry", s.loaddeliveryentry)
	router.POST("/mixingcolor/senddeliveryentry", s.senddeliveryentry)
	router.GET("/mixingcolor/batchentry", s.mc_batchentry)
	router.GET("/mixingcolor/entry/loadbatchform", s.mce_loadbatchform)
	router.POST("/mixingcolor/entry/sendbatchentry", s.mce_sendbatchentry)

	router.POST("/mixingcolor/mixingfilter", s.mixingfilter)

	router.POST("/mixingcolor/mixingreports/:batchno", s.mixingreports)
	router.GET("/mixingcolor/deletereports", s.deletereport)

	router.GET("/mixingcolor/addcolorform", s.addcolorform)
	router.PUT("/mixingcolor/addcolor", s.addcolor)
	router.POST("/mixingcolor/colorsearch", s.colorsearch)
	router.POST("/mixingcolor/colorfilter", s.colorfilter)

	router.GET("/mixingcolor/usingentry", s.mc_usingreports)
	router.GET("/mixingcolor/entry/loadusingform", s.mc_loadusingform)
	router.GET("/mixingcolor/entry/getupdateform", s.mc_getupdateform)
	router.POST("/mixingcolor/entry/updateusingtime", s.mc_updateusingtime)
	router.POST("/mixingcolor/entry/updateusingitem", s.mc_updateusingitem)
	router.POST("/mixingcolor/entry/searchcolorcode", s.mce_searchcolorcode)
	router.POST("/mixingcolor/entry/createfastbatch", s.mce_createfastbatch)

	router.GET("/mixingcolor/colorentry", s.mc_colorentry)
	router.GET("/mixingcolor/entry/loadcolorform", s.mc_loadcolorform)

	router.GET("/colormixing/overview", s.c_overview)
	router.GET("/colormixing/overview/changedisplay/:type", s.co_changedisplay)
	router.POST("/colormixing/overview/searchbatch", s.co_searchbatch)
	// router.POST("/colormixing/overview/filterbatch", s.co_filterbatch)
	// router.POST("/colormixing/overview/batchitems/:batchno", s.co_batchitems)
	// router.POST("/colormixing/overview/searchcolor", s.co_searchcolor)
	// router.POST("/colormixing/overview/filtercolor", s.co_filtercolor)
	// router.POST("/colormixing/overview/createstandard", s.co_createstandard)
	// router.DELETE("/colormixing/overview/deletestandard/:name", s.co_deletestandard)

	// router.GET("/colormixing/admin", withAuth(s.c_admin))
	router.GET("/colormixing/admin/loadbatchentry", s.ca_loadbatchentry)
	// router.POST("/colormixing/admin/sendbatchentry", s.ca_sendmixingentry)
	router.GET("/colormixing/admin/loadpanelentry", s.ca_loadpanelentry)
	// router.POST("/colormixing/admin/sendpanelentry", s.ca_sendpanelentry)
	router.DELETE("/colormixing/admin/deletepanel/:id", s.ca_deletepanel)
	router.GET("/colormixing/admin/panelupdateform/:id", s.ca_panelupdateform)
	router.PUT("/colormixing/admin/updatepanel/:id", s.ca_updatepanel)
	router.GET("/colormixing/admin/loadusingtimeform", s.ca_loadusingtimeform)
	router.GET("/colormixing/admin/loadinspectionform", s.ca_loadinspectionform)
	// router.POST("/colormixing/admin/addinspection", s.ca_addinspection)
	router.GET("/colormixing/admin/loadmixingbatch", s.ca_loadmixingbatch)
	router.POST("/colormixing/admin/searchbatch", s.ca_searchbatch)
	// router.GET("/colormixing/admin/loadcolorpanel", s.ca_loadcolorpanel)
	// router.POST("/colormixing/admin/searchpanel", s.ca_searchpanel)
	router.DELETE("/colormixing/admin/deletemixing/:batchno", s.ca_deletemixing)
	router.GET("/colormixing/admin/batchupdateform/:batchno", s.ca_batchupdateform)
	// router.PUT("/colormixing/admin/updatebatch/:batchno", s.ca_updatebatch)
	router.GET("/colormixing/admin/loadauditentry", s.ca_loadauditentry)
	router.POST("/colormixing/admin/failaudit/:id", s.ca_failaudit)
	router.POST("/colormixing/admin/passaudit/:id", s.ca_passaduti)

	////////////////////////////////////////////////////////////////////
	// Quality ////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////
	router.GET("/quality/fastentry", withAuth(s.q_fastentry))
	router.GET("/quality/entry/loadform", s.q_loadform)
	router.POST("/quality/entry/sendentry", s.q_sendentry)
	// end Quality

	router.GET("/gnhh/overview", s.g_overview)
	router.GET("/gnhh/overview/loadchart", s.go_loadchart)
	router.GET("/gnhh/overview/loadproducttree", s.go_loadproducttree)
	router.GET("/gnhh/overview/loadtimeline", s.go_loadtimeline)
	router.GET("/gnhh/overview/loaddetail", s.go_loaddetail)
	router.POST("/gnhh/overview/updatetimeline", s.go_updatetimeline)
	router.POST("/gnhh/overview/searchtimeline", s.go_searchtimeline)
	router.POST("/gnhh/overview/filtertimeline", s.go_filtertimeline)
	router.POST("/gnhh/overview/iteminfo", s.go_iteminfo)
	router.GET("/gnhh/overview/loadtree", s.go_loadtree)
	router.POST("/gnhh/overview/getproductcodes", s.go_getproductcodes)
	router.POST("/gnhh/overview/mofilter", s.go_mofilter)
	router.POST("/gnhh/overview/productfilter", s.go_productfilter)
	router.POST("/gnhh/overview/searchdetail", s.go_searchdetail)

	router.GET("/gnhh/entry/import", s.ge_import)
	router.POST("/gnhh/entry/importdata", s.ge_importdata)

	router.GET("/safety/entry", s.s_entry)
	router.POST("/safety/sendentry", s.s_sendentry)

	router.POST("/manpower/importfile", s.mp_importfile)
	router.POST("/otattend/importfile", s.oa_importfile)

	router.GET("/mo/entry", withAuth(s.mo_entry))
	router.POST("/mo/entry", s.mo_insertMoList)

	router.GET("/mo/admin", s.mo_admin)

	router.GET("/item/entry", withAuth(s.i_entry))
	router.POST("/item/entry", s.i_importitemlist)

	router.GET("/item/admin", s.i_admin)
	// router.POST("/item/admin/additem", s.i_additem)
	router.POST("/item/admin/addpart", s.i_addpart)

	router.GET("/test", s.handleGetTest)

	router.GET("/test/loadmain", s.testload)

	return router
}
func withAuth(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		authurlsToken, err := r.Cookie("authurls")
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if !strings.Contains(authurlsToken.Value, r.URL.Path) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized) // should design a page for unauthorized request
			return
		}
		h(w, r, ps)
	}
}

func withAuth2(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		authurlsToken, err := r.Cookie("authurls")
		if err != nil {
			log.Println(err)
			w.Write([]byte("Not log in yet"))
			return
		}

		if !strings.Contains(authurlsToken.Value, r.URL.Path) {
			w.Write([]byte("không có thẩm quyền"))
			return
		}
		h(w, r, ps)
	}
}
func (s *Server) writeJSON(w http.ResponseWriter, data interface{}, status int) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (s *Server) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func OpenPgDB(conStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

type Validator struct {
	Errors map[string]string
}

func NewValidator() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		if _, exists := uniqueValues[value]; exists {
			return false
		}
		uniqueValues[value] = true
	}
	return true
}

func (s *Server) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := map[string]interface{}{
		"error": message,
	}

	if err := s.writeJSON(w, env, status); err != nil {
		s.Logger.Println(err)
		w.WriteHeader(500)
	}
}

func (s *Server) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	s.Logger.Println(err)
	message := "The server encountered an error and could not process your request."
	s.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (s *Server) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found."
	s.errorResponse(w, r, http.StatusNotFound, message)
}

func (s *Server) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not allowed for this resource.", r.Method)
	s.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (s *Server) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	s.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (s *Server) failedValidationResponse(w http.ResponseWriter, r *http.Request, err map[string]string) {
	s.errorResponse(w, r, http.StatusUnprocessableEntity, err)
}

// //////////////////////////////////////////////////////////
// / - Get index page
// //////////////////////////////////////////////////////////
func (s *Server) index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defaulturlToken, err := r.Cookie("defaulturl")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if defaulturlToken.Value != "" {
		http.Redirect(w, r, defaulturlToken.Value, http.StatusSeeOther)
		return
	}

	usernameToken, err := r.Cookie("username") // check for username on cookie, not found, go to login page
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user = User{} // if sessionToken found, check token with database, get the right page for that user
	if err = s.mgdb.Collection("user").FindOne(context.Background(), bson.M{"username": usernameToken.Value}).Decode(&user); err != nil {
		s.Logger.Println("Fail to decode defaulturl", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if user.Defaulturl == "" { // go to home page when default url is missing in database
		user.Defaulturl = "/home"
	}

	http.Redirect(w, r, user.Defaulturl, http.StatusSeeOther) // go to user's default page
}

// //////////////////////////////////////////////////////////
// /home - Get
// //////////////////////////////////////////////////////////
func (s *Server) home(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/home/home.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////
// /login - Get
// //////////////////////////////////////////////////////////
func (s *Server) serveLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data := map[string]string{
		"msg": "Login as guest if you do not have account. Want an account, click Request",
	}

	template.Must(template.ParseFiles("templates/pages/login/login.html")).Execute(w, data)
}

// //////////////////////////////////////////////////////////
// /login - Post
// //////////////////////////////////////////////////////////
func (s *Server) requestLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" && password == "" {
		username = "guest"
		password = "guest"
	}
	user := User{}

	if err := s.mgdb.Collection("user").FindOne(context.Background(), bson.M{"username": username}).Decode(&user); err != nil {
		log.Println(err)
		data := map[string]string{
			"msg": "Username is incorrect, plaese check again. Do not have? click Request",
		}
		template.Must(template.ParseFiles("templates/pages/login/login.html")).Execute(w, data)
		return
	}

	if user.Password != password {
		data := map[string]string{
			"msg": "Password is incorrect, plaese check again. Forgot? Click Request",
		}
		template.Must(template.ParseFiles("templates/pages/login/login.html")).Execute(w, data)
		return
	}

	// when username and password are correct
	http.SetCookie(w, &http.Cookie{
		Name:    "username",
		Value:   user.Username,
		Expires: time.Now().Add(720 * time.Hour),
		Path:    "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "staffid",
		Value:   user.Info.StaffId,
		Expires: time.Now().Add(720 * time.Hour),
		Path:    "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "defaulturl",
		Value:   user.Defaulturl,
		Expires: time.Now().Add(720 * time.Hour),
		Path:    "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "authurls",
		Value:   strings.Join(user.Authurls, " "),
		Expires: time.Now().Add(720 * time.Hour),
		Path:    "/",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// //////////////////////////////////////////////////////////
// /logout
// //////////////////////////////////////////////////////////
func (s *Server) logout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	http.SetCookie(w, &http.Cookie{
		Name:    "username",
		Value:   "",
		Expires: time.Now(),
		Path:    "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "defaulturl",
		Value:   "",
		Expires: time.Now(),
		Path:    "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "permission",
		Value:   "",
		Expires: time.Now(),
		Path:    "/",
	})

	data := map[string]string{
		"msg": "Logout successful! For more information, click Request and send a request to admin",
	}
	template.Must(template.ParseFiles("templates/pages/login/login.html")).Execute(w, data)
}

// //////////////////////////////////////////////////////////
// /admin
// //////////////////////////////////////////////////////////
func (s *Server) admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////
// /dashboard
// //////////////////////////////////////////////////////////
func (s *Server) dashboard(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// get data for cutting chart
	// pipeline := mongo.Pipeline{
	// 	{{"$match", bson.M{"type": "report", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -15))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
	// 	{{"$addFields", bson.M{"is25": bson.M{"$eq": bson.A{"$thickness", 25}}}}},
	// 	{{"$group", bson.M{"_id": bson.M{"date": "$date", "is25": "$is25"}, "qty": bson.M{"$sum": "$qtycbm"}}}},
	// 	{{"$sort", bson.D{{"_id.date", 1}, {"_id.is25", 1}}}},
	// 	{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "is25": "$_id.is25"}}},
	// 	{{"$unset", "_id"}},
	// }
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"type": "report", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -18))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "is25reeded": "$is25reeded"}, "qty": bson.M{"$sum": "$qtycbm"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.is25reeded", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "is25": "$_id.is25reeded"}}},
		{{"$unset", "_id"}},
	}

	cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Println(err)
		return
	}
	defer cur.Close(context.Background())
	var cuttingData []struct {
		Date string  `bson:"date" json:"date"`
		Is25 bool    `bson:"is25" json:"is25"`
		Qty  float64 `bson:"qty" json:"qty"`
	}

	if err = cur.All(context.Background(), &cuttingData); err != nil {
		log.Println(err)
		return
	}

	//get wood return
	cur, err = s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "return", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -18))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "is25": "$is25"}, "qty": bson.M{"$sum": "$qtycbm"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.is25", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "is25": "$_id.is25"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
		return
	}
	var cuttingReturnData []struct {
		Date string  `bson:"date" json:"date"`
		Is25 bool    `bson:"is25" json:"is25"`
		Qty  float64 `bson:"qty" json:"qty"`
	}
	if err := cur.All(context.Background(), &cuttingReturnData); err != nil {
		log.Println(err)
		return
	}

	//get fine wood
	cur, err = s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "fine", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -18))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "is25reeded": "$is25reeded"}, "qty": bson.M{"$sum": "$qtycbm"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.is25reeded", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "is25reeded": "$_id.is25reeded"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
		return
	}
	var cuttingFineData []struct {
		Date       string  `bson:"date" json:"date"`
		Is25reeded bool    `bson:"is25reeded" json:"is25reeded"`
		Qty        float64 `bson:"qty" json:"qty"`
	}
	if err := cur.All(context.Background(), &cuttingFineData); err != nil {
		log.Println(err)
		return
	}

	//get target data for leftchart
	sr := s.mgdb.Collection("cutting").FindOne(context.Background(), bson.M{"type": "target"}, options.FindOne().SetSort(bson.M{"startdate": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var targetactualData struct {
		Name      string    `bson:"name" json:"name"`
		StartDate time.Time `bson:"startdate"`
		EnddDate  time.Time `bson:"enddate"`
		Detail    []struct {
			Type   string  `bson:"type" json:"type"`
			Target float64 `bson:"target" json:"target"`
		} `bson:"detail" json:"detail"`
		StartDateStr string `json:"startdate"`
		EndDateStr   string `json:"enddate"`
	}
	if err := sr.Decode(&targetactualData); err != nil {
		log.Println(err)
	}
	targetactualData.StartDateStr = targetactualData.StartDate.Format("02/01/2006")
	targetactualData.EndDateStr = targetactualData.EnddDate.Format("02/01/2006")

	cur, err = s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": "fine"}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(targetactualData.StartDate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(targetactualData.EnddDate)}}}}}},
		{{"$set", bson.M{"is25reeded": bson.M{"$ifNull": bson.A{"$is25reeded", false}}}}},
		{{"$group", bson.M{"_id": "$is25reeded", "qty": bson.M{"$sum": "$qtycbm"}}}},
		{{"$sort", bson.D{{"_id", 1}}}},
		{{"$set", bson.M{"prodtype": "$_id"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer cur.Close(context.Background())
	var cuttingProdtypeData []struct {
		Prodtype bool    `bson:"prodtype" json:"prodtype"`
		Qty      float64 `bson:"qty" json:"qty"`
	}

	if err = cur.All(context.Background(), &cuttingProdtypeData); err != nil {
		log.Println(err)
		return
	}

	if len(cuttingProdtypeData) == 1 {
		if cuttingProdtypeData[0].Prodtype {
			cuttingProdtypeData = append(cuttingProdtypeData, struct {
				Prodtype bool    `bson:"prodtype" json:"prodtype"`
				Qty      float64 `bson:"qty" json:"qty"`
			}{
				Prodtype: false, Qty: 0,
			})
		} else {
			cuttingProdtypeData = append(cuttingProdtypeData, struct {
				Prodtype bool    `bson:"prodtype" json:"prodtype"`
				Qty      float64 `bson:"qty" json:"qty"`
			}{
				Prodtype: true, Qty: 0,
			})
		}
	}

	//get target line data of cutting
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"name": "cutting total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -18))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var cuttingTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &cuttingTarget); err != nil {
		log.Println(err)
	}

	// get last update time of cutting
	cuttingSr := s.mgdb.Collection("cutting").FindOne(context.Background(), bson.M{"type": "report"}, options.FindOne().SetSort(bson.M{"createddate": -1}))
	if cuttingSr.Err() != nil {
		log.Println(cuttingSr.Err())
	}
	var cuttingLastReport struct {
		CreatedDate time.Time `bson:"createddate" json:"createddate"`
	}
	if err := cuttingSr.Decode(&cuttingLastReport); err != nil {
		log.Println(err)
	}
	cuttingUpTime := cuttingLastReport.CreatedDate.Add(7 * time.Hour).Format("15:04")

	// get data for lamination
	cur, err = s.mgdb.Collection("lamination").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "prodtype": "$prodtype"}, "qty": bson.M{"$sum": "$qty"}}}},
		{{"$sort", bson.M{"_id.date": 1, "_id.prodtype": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "prodtype": "$_id.prodtype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	var laminationChartData []struct {
		Date     string  `bson:"date" json:"date"`
		Prodtype string  `bson:"prodtype" json:"prodtype"`
		Qty      float64 `bson:"qty" json:"qty"`
	}
	if err := cur.All(context.Background(), &laminationChartData); err != nil {
		log.Println(err)
	}
	// get target of lamination
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"name": "lamination total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var laminationTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &laminationTarget); err != nil {
		log.Println(err)
	}
	// get last update time of lamination
	laminationSr := s.mgdb.Collection("lamination").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if cuttingSr.Err() != nil {
		log.Println(cuttingSr.Err())
	}
	var laminationLastReport struct {
		CreatedDate time.Time `bson:"createdat" json:"createdat"`
	}
	if err := laminationSr.Decode(&laminationLastReport); err != nil {
		log.Println(err)
	}
	laminationUpTime := laminationLastReport.CreatedDate.Add(7 * time.Hour).Format("15:04")

	// get data for Packing Chart
	cur, err = s.mgdb.Collection("packchart").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"of": "packchart"}}},
		{{"$sort", bson.M{"date": -1}}},
		{{"$limit", 20}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println("failed to get data from packchart", err)
	}
	var packchartData []struct {
		Date   string  `bson:"date" json:"date"`
		Brand1 float64 `bson:"1-brand" json:"1-brand"`
		Brand2 float64 `bson:"2-brand" json:"2-brand"`
		Rh1    float64 `bson:"1-rh" json:"1-rh"`
		Rh2    float64 `bson:"2-rh" json:"2-rh"`
	}
	if err := cur.All(context.Background(), &packchartData); err != nil {
		log.Println("failed to decode", err)
	}

	template.Must(template.ParseFiles(
		"templates/pages/dashboard/dashboard.html",
		"templates/pages/dashboard/cuttingchart.html",
		"templates/pages/dashboard/laminationchart.html",
		"templates/pages/dashboard/packingchart.html",
		"templates/pages/dashboard/provalcht.html",
		"templates/shared/navbar.html",
	)).Execute(w, map[string]interface{}{
		"cuttingData":         cuttingData,
		"cuttingReturnData":   cuttingReturnData,
		"cuttingFineData":     cuttingFineData,
		"targetactualData":    targetactualData,
		"cuttingProdtypeData": cuttingProdtypeData,
		"cuttingTarget":       cuttingTarget,
		"cuttingUpTime":       cuttingUpTime,
		"laminationChartData": laminationChartData,
		"laminationTarget":    laminationTarget,
		"laminationUpTime":    laminationUpTime,
		"packingData":         packchartData,
	})
}

// //////////////////////////////////////////////////////////
// router.GET("/dashboard/loadotattend", s.d_loadotattend)
// //////////////////////////////////////////////////////////
func (s *Server) d_loadotattend(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////
// router.GET("/dashboard/loadmanpower", s.d_loadmanpower)
// //////////////////////////////////////////////////////////
func (s *Server) d_loadmanpower(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////
// router.POST("/dashboard/loadproductionvop", s.d_loadproductionvop)
// //////////////////////////////////////////////////////////
func (s *Server) d_loadproductionvop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -16))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": "$date", "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.M{"_id": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var productiondata []struct {
		Date  string  `json:"date"`
		Value float64 `json:"value"`
	}

	if err := cur.All(context.Background(), &productiondata); err != nil {
		log.Println(err)
	}

	cur, err = s.mgdb.Collection("vopmanhr").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -16))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var manhrdata []struct {
		Date  string  `bson:"date" json:"date"`
		Manhr float64 `bson:"manhr" json:"manhr"`
	}

	if err := cur.All(context.Background(), &manhrdata); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/dashboard/productionvopchart.html")).Execute(w, map[string]interface{}{
		"productiondata": productiondata,
		"manhrdata":      manhrdata,
	})
}

// //////////////////////////////////////////////////////////
// router.GET("/dashboard/loadplan", s.d_loadproductionplan)
// //////////////////////////////////////////////////////////
func (s *Server) d_loadproductionplan(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////
// router.GET("/dashboard/loadrawwood", s.d_loadrawwood)
// //////////////////////////////////////////////////////////
func (s *Server) d_loadrawwood(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// get data of raw wood input
	cur, err := s.mgdb.Collection("rawwood").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": "import"}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -15))}}}}}},
		{{"$group", bson.M{"_id": "$date", "qty": bson.M{"$sum": "$qty"}}}},
		{{"$sort", bson.M{"_id": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var rawwoodImportData []struct {
		Date string  `bson:"date" json:"date"`
		Qty  float64 `bson:"qty" json:"qty"`
	}
	if err := cur.All(context.Background(), &rawwoodImportData); err != nil {
		log.Println(err)
	}

	// get data of selection
	cur, err = s.mgdb.Collection("rawwood").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": "selection"}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -15))}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "woodtone": "$woodtone"}, "qty": bson.M{"$sum": "$qty"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.woodtone", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "woodtone": "$_id.woodtone"}}},
		{{"$unset", bson.A{"_id.date", "_id.woodtone"}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var rawwoodSelectionData []struct {
		Date     string  `bson:"date" json:"date"`
		Woodtone string  `bson:"woodtone" json:"woodtone"`
		Qty      float64 `bson:"qty" json:"qty"`
	}
	if err := cur.All(context.Background(), &rawwoodSelectionData); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/dashboard/rawwood.html")).Execute(w, map[string]interface{}{
		"rawwoodImportData":    rawwoodImportData,
		"rawwoodSelectionData": rawwoodSelectionData,
	})
}

// //////////////////////////////////////////////////////////
// /dashboard/loadproduction - load production area in dashboard
// //////////////////////////////////////////////////////////
func (s *Server) d_loadproduction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pvPipeline := mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype", "item": "$item"}, "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.M{"_id.date": -1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "factory": bson.M{"$concat": bson.A{"Factory ", "$_id.factory"}}, "type": bson.M{"$toUpper": "$_id.prodtype"}, "item": "$_id.item"}}},
		{{"$unset", "_id"}},
	}
	cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), pvPipeline)
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	type ProdValue struct {
		Date    string  `json:"date"`
		Factory string  `json:"factory"`
		Type    string  `json:"prodtype"`
		Item    string  `json:"item"`
		Value   float64 `json:"value"`
	}
	var productiondata []ProdValue

	if err := cur.All(context.Background(), &productiondata); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/dashboard/productionchart.html")).Execute(w, map[string]interface{}{
		"productiondata": productiondata,
	})
}

// //////////////////////////////////////////////////////////////
// router.POST("/dashboard/otattend/getchart", s.doa_getchart)
// //////////////////////////////////////////////////////////////
func (s *Server) doa_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////////
// router.POST("/dashboard/manpower/getchart", s.dmp_getchart)
// //////////////////////////////////////////////////////////////
func (s *Server) dmp_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	days, _ := strconv.Atoi(r.FormValue("numberofdays"))
	cur, err := s.mgdb.Collection("manpower").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"name": -1}).SetLimit(int64(days)))
	if err != nil {
		log.Println(err)
	}

	type MP struct {
		Name     string                   `bson:"name" json:"name"`
		Children []map[string]interface{} `bson:"children" json:"children"`
	}

	var ss []MP
	if err := cur.All(context.Background(), &ss); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/dashboard/manpower_statuschart.html")).Execute(w, map[string]interface{}{
		"manpowerdata": ss,
	})
}

// ////////////////////////////////////////////////////////////////////////////////
//
//	router.POST("/dashboard/productionvop/getchart", s.dpv_getchart)
//
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dpv_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("vopcharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("vopFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("vopToDate"))

	switch pickedChart {
	case "value-man":
		cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())

		var productiondata []struct {
			Date  string  `json:"date"`
			Value float64 `json:"value"`
		}

		if err := cur.All(context.Background(), &productiondata); err != nil {
			log.Println(err)
		}

		cur, err = s.mgdb.Collection("vopmanhr").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())

		var manhrdata []struct {
			Date  string  `bson:"date" json:"date"`
			Manhr float64 `bson:"manhr" json:"manhr"`
		}

		if err := cur.All(context.Background(), &manhrdata); err != nil {
			log.Println(err)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/productionvop_genchart.html")).Execute(w, map[string]interface{}{
			"productiondata": productiondata,
			"manhrdata":      manhrdata,
		})

	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/production/getchart - change chart of production area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dpr_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("productioncharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("productionFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("productionToDate"))

	switch pickedChart {
	case "value":
		pvPipeline := mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype", "item": "$item"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.M{"_id.date": -1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "factory": bson.M{"$concat": bson.A{"Factory ", "$_id.factory"}}, "type": bson.M{"$toUpper": "$_id.prodtype"}, "item": "$_id.item"}}},
			{{"$unset", "_id"}},
		}
		cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), pvPipeline)
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		type ProdValue struct {
			Date    string  `json:"date"`
			Factory string  `json:"factory"`
			Type    string  `json:"prodtype"`
			Item    string  `json:"item"`
			Value   float64 `json:"value"`
		}
		var productiondata []ProdValue

		if err := cur.All(context.Background(), &productiondata); err != nil {
			log.Println(err)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/prod_value.html")).Execute(w, map[string]interface{}{
			"productiondata": productiondata,
		})

	case "mtd":
		mtds, _ := strconv.Atoi(r.FormValue("numberOfMTDs"))
		mtdFromDate := time.Date(time.Now().Year(), time.Now().Month()-time.Month(mtds), 1, 0, 0, 0, 0, time.Now().Location())
		cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(mtdFromDate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$group", bson.M{"_id": bson.M{"$month": "$_id"}, "value": bson.M{"$push": "$value"}}}},
			{{"$set", bson.M{"month": "$_id"}}},
			{{"$unset", "_id"}},
			{{"$sort", bson.M{"month": 1}}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var resu []struct {
			Month int       `bson:"month" json:"month"`
			Value []float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &resu); err != nil {
			log.Println(err)
		}

		type PP struct {
			Month string `json:"month"`
			Data  []struct {
				Days     int     `json:"days"`
				AccValue float64 `json:"value"`
			} `json:"dat"`
		}

		var productiondata []PP
		for _, re := range resu {
			var a PP
			a.Month = time.Month(re.Month).String()
			for i := 0; i < len(re.Value); i++ {
				if i == 0 {
					a.Data = append(a.Data, struct {
						Days     int     `json:"days"`
						AccValue float64 `json:"value"`
					}{Days: i + 1, AccValue: re.Value[i]})
				} else {
					a.Data = append(a.Data, struct {
						Days     int     `json:"days"`
						AccValue float64 `json:"value"`
					}{Days: i + 1, AccValue: a.Data[i-1].AccValue + re.Value[i]})
				}
			}
			productiondata = append(productiondata, a)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/prod_mtd.html")).Execute(w, map[string]interface{}{
			"productiondata": productiondata,
		})
	}
}

// /////////////////////////////////////////////////////////////
// /dashboard/loadreededline - load reededline area in dashboard
// /////////////////////////////////////////////////////////////
func (s *Server) d_loadreededline(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("reededline").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "tone": "$tone"}, "qty": bson.M{"$sum": "$qty"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.tone", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "tone": "$_id.tone"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var reededlinedata []struct {
		Date string  `bson:"date" json:"date"`
		Tone string  `bson:"tone" json:"tone"`
		Qty  float64 `bson:"qty" json:"qty"`
	}
	if err := cur.All(context.Background(), &reededlinedata); err != nil {
		log.Println(err)
	}

	// get data of Gỗ 25 of cutting
	cur, err = s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"is25reeded": true}, bson.M{"type": "report"}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": "$date", "qty": bson.M{"$sum": "$qtycbm"}}}},
		{{"$sort", bson.M{"_id": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	var wood25data []struct {
		Date string  `bson:"date" json:"date"`
		Qty  float64 `bson:"qty" json:"qty"`
	}
	if err := cur.All(context.Background(), &wood25data); err != nil {
		log.Println(err)
	}

	// get target of reededline
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"name": "reededline total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var reededlineTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &reededlineTarget); err != nil {
		log.Println(err)
	}

	// get time of latest update
	sr := s.mgdb.Collection("reededline").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var LastReport struct {
		Createdat time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&LastReport); err != nil {
		log.Println(err)
	}
	reededlineUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

	template.Must(template.ParseFiles("templates/pages/dashboard/reededline.html")).Execute(w, map[string]interface{}{
		"reededlinedata":   reededlinedata,
		"wood25data":       wood25data,
		"reededlineTarget": reededlineTarget,
		"reededlineUpTime": reededlineUpTime,
	})
}

// /////////////////////////////////////////////////////////////
// /dashboard/loadreededoutput - load reededoutput area in dashboard
// /////////////////////////////////////////////////////////////
func (s *Server) d_loadoutput(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("output").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "reeded"}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "section": "$section"}, "qty": bson.M{"$sum": "$qty"}, "type": bson.M{"$first": "$type"}}}},
		{{"$set", bson.M{"section": "$_id.section", "date": "$_id.date"}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$group", bson.M{"_id": "$section", "type": bson.M{"$first": "$type"}, "qty": bson.M{"$sum": "$qty"}, "avg": bson.M{"$avg": "$qty"}, "lastdate": bson.M{"$last": "$date"}}}},
		{{"$sort", bson.M{"_id": 1}}},
		{{"$set", bson.M{"section": bson.M{"$substr": bson.A{"$_id", 2, -1}}, "lastdate": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$lastdate"}}}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var reededoutputData []struct {
		Section   string  `bson:"section" json:"section"`
		Type      string  `bson:"type" json:"type"`
		Qty       float64 `bson:"qty" json:"qty"`
		Avg       float64 `bson:"avg" json:"avg"`
		LastDate  string  `bson:"lastdate" json:"lastdate"`
		Inventory float64 `bson:"inventory" json:"inventory"`
	}
	if err := cur.All(context.Background(), &reededoutputData); err != nil {
		log.Println(err)
	}
	// get latest inventory
	sr := s.mgdb.Collection("output").FindOne(context.Background(), bson.M{"section": "a.Inventory"}, options.FindOne().SetSort(bson.M{"date": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var latestInventory struct {
		Date    time.Time `bson:"date"`
		Section string    `json:"section"`
		Qty     float64   `bson:"qty" json:"qty"`
		DateStr string    `json:"date"`
	}
	if err := sr.Decode(&latestInventory); err != nil {
		log.Println(err)
	}
	latestInventory.Section = "Inventory"
	latestInventory.DateStr = latestInventory.Date.Format("02-01-2006")
	// get last update time
	sr = s.mgdb.Collection("output").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var latestOne struct {
		Date time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&latestOne); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/dashboard/output.html")).Execute(w, map[string]interface{}{
		"reededoutputData": reededoutputData,
		"latestInventory":  latestInventory,
		"outputUpTime":     latestOne.Date.Add(7 * time.Hour).Format("15:04 ngày 02-01-2006"),
	})
}

// //////////////////////////////////////////////////////////
// /dashboard/loadpanelcnc - load panelcnc area in dashboard
// //////////////////////////////////////////////////////////
func (s *Server) d_loadpanelcnc(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, 1))}}}}}},
		{{"$group", bson.M{"_id": "$date", "qty": bson.M{"$sum": "$qty"}}}},
		{{"$sort", bson.D{{"_id", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
		{{"$unset", "_id"}},
	}
	cur, err := s.mgdb.Collection("panelcnc").Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var panelChartData []struct {
		Date string  `bson:"date" json:"date"`
		Qty  float64 `bson:"qty" json:"qty"`
	}

	if err := cur.All(context.Background(), &panelChartData); err != nil {
		log.Println(err)
	}
	// get target of panelcnc
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"name": "panelcnc total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var panelcncTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &panelcncTarget); err != nil {
		log.Println(err)
	}

	// get time of latest update
	sr := s.mgdb.Collection("panelcnc").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var LastReport struct {
		Createdat time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&LastReport); err != nil {
		log.Println(err)
	}
	panelcncUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

	template.Must(template.ParseFiles("templates/pages/dashboard/panelcncchart.html")).Execute(w, map[string]interface{}{
		"panelChartData": panelChartData,
		"panelcncTarget": panelcncTarget,
		"panelcncUpTime": panelcncUpTime,
	})
}

// //////////////////////////////////////////////////////////
// /dashboard/loadveneer - load veneer area in dashboard
// //////////////////////////////////////////////////////////
func (s *Server) d_loadveneer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("veneer").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -15))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "type": "$type"}, "qty": bson.M{"$sum": "$qty"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.type", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": "$_id.type"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var veneerChartData []struct {
		Date string  `bson:"date" json:"date"`
		Type string  `bson:"type" json:"type"`
		Qty  float64 `bson:"qty" json:"qty"`
	}
	if err := cur.All(context.Background(), &veneerChartData); err != nil {
		log.Println(err)
	}
	// get target for veneer
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"name": "veneer total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var veneerTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &veneerTarget); err != nil {
		log.Println(err)
	}

	// get time of latest update
	sr := s.mgdb.Collection("veneer").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var LastReport struct {
		Createdat time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&LastReport); err != nil {
		log.Println(err)
	}
	veneerUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

	template.Must(template.ParseFiles("templates/pages/dashboard/veneer.html")).Execute(w, map[string]interface{}{
		"veneerChartData": veneerChartData,
		"veneerTarget":    veneerTarget,
		"veneerUpTime":    veneerUpTime,
	})
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/loadassembly - load assembly area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) d_loadassembly(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.factory", 1}, {"_id.prodtype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var assemblyData []struct {
		Date  string  `bson:"date" json:"date"`
		Type  string  `bson:"type" json:"type"`
		Value float64 `bson:"value" json:"value"`
	}
	if err := cur.All(context.Background(), &assemblyData); err != nil {
		log.Println(err)
	}

	// get plan data
	cur, err = s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
		// {{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -12))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}}}}},
		{{"$sort", bson.M{"createdat": -1}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "plantype": "$plantype"}, "plan": bson.M{"$first": "$plan"}, "plans": bson.M{"$firstN": bson.M{"input": "$plan", "n": 2}}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.plantype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "plantype": "$_id.plantype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var assemblyPlanData []struct {
		Date     string    `bson:"date" json:"date"`
		Plantype string    `bson:"plantype" json:"plantype"`
		Plan     float64   `bson:"plan" json:"plan"`
		Plans    []float64 `bson:"plans" json:"plans"`
		Change   float64   `json:"change"`
	}

	if err := cur.All(context.Background(), &assemblyPlanData); err != nil {
		log.Println(err)
	}
	for i := 0; i < len(assemblyPlanData); i++ {
		if len(assemblyPlanData[i].Plans) >= 2 && assemblyPlanData[i].Plans[1] != 0 {
			assemblyPlanData[i].Change = assemblyPlanData[i].Plans[1] - assemblyPlanData[i].Plan
		} else {
			assemblyPlanData[i].Change = 0
		}
	}

	// get inventory
	cur, err = s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "Inventory", "factory": bson.M{"$exists": true}}}},
		{{"$sort", bson.M{"createdat": -1}}},
		{{"$limit", 20}},
		{{"$group", bson.M{"_id": bson.M{"factory": "$factory", "prodtype": "$prodtype"}, "inventory": bson.M{"$first": "$inventory"}, "createdat": bson.M{"$first": "$createdat"}}}},
		{{"$set", bson.M{"type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
		{{"$sort", bson.M{"type": 1}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var assemblyInventoryData []struct {
		Type         string    `bson:"type" json:"type"`
		Inventory    float64   `bson:"inventory" json:"inventory"`
		CreatedAt    time.Time `bson:"createdat"`
		CreatedAtStr string    `json:"createdat"`
	}
	if err := cur.All(context.Background(), &assemblyInventoryData); err != nil {
		log.Println(err)
	}

	for i := 0; i < len(assemblyInventoryData); i++ {
		assemblyInventoryData[i].CreatedAtStr = assemblyInventoryData[i].CreatedAt.Add(7 * time.Hour).Format("15h04 date 2/1")
	}

	// get target
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		// {{"$match", bson.M{"name": "assembly total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$match", bson.M{"name": "assembly total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var assemblyTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &assemblyTarget); err != nil {
		log.Println(err)
	}

	// get time of latest update
	sr := s.mgdb.Collection("assembly").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var LastReport struct {
		Createdat time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&LastReport); err != nil {
		log.Println(err)
	}
	assemblyUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")
	template.Must(template.ParseFiles("templates/pages/dashboard/assembly.html")).Execute(w, map[string]interface{}{
		"assemblyData":          assemblyData,
		"assemblyPlanData":      assemblyPlanData,
		"assemblyInventoryData": assemblyInventoryData,
		"assemblyTarget":        assemblyTarget,
		"assemblyUpTime":        assemblyUpTime,
	})
}

// router.GET("/dashboard/loadwhitewood", s.d_loadwhitewood)
func (s *Server) d_loadwhitewood(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("whitewood").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -12))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.prodtype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": "$_id.prodtype"}}},
		// {{"$set", bson.M{"date": "_id.date", "type": "$_id.prodtype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var whitewoodData []struct {
		Date  string  `bson:"date" json:"date"`
		Type  string  `bson:"type" json:"type"`
		Value float64 `bson:"value" json:"value"`
	}
	if err := cur.All(context.Background(), &whitewoodData); err != nil {
		log.Println(err)
	}

	// get avg of this month
	fromdate := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	todate := fromdate.AddDate(0, 1, 0)
	cur, err = s.mgdb.Collection("whitewood").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$group", bson.M{"_id": "$date", "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.M{"_id": 1}}},
	})
	if err != nil {
		log.Println(err)
	}
	var valuedata []struct {
		Date  time.Time `bson:"_id"`
		Value float64   `bson:"value" json:"value"`
	}
	if err := cur.All(context.Background(), &valuedata); err != nil {
		log.Println(err)
	}

	if len(valuedata) > 0 && valuedata[len(valuedata)-1].Date.Format("2006-01-02") == time.Now().Format("2006-01-02") {
		valuedata = valuedata[:len(valuedata)-1]
	}
	var total float64
	for _, v := range valuedata {
		total += v.Value
	}
	avgdata := total / float64(len(valuedata))

	// get plan data
	cur, err = s.mgdb.Collection("whitewood").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -12))}}}}}},
		{{"$sort", bson.M{"createdat": -1}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "plantype": "$plantype"}, "plan": bson.M{"$first": "$plan"}, "plans": bson.M{"$firstN": bson.M{"input": "$plan", "n": 2}}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.plantype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "plantype": "$_id.plantype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var whitewoodPlanData []struct {
		Date     string    `bson:"date" json:"date"`
		Plantype string    `bson:"plantype" json:"plantype"`
		Plan     float64   `bson:"plan" json:"plan"`
		Plans    []float64 `bson:"plans" json:"plans"`
		Change   float64   `json:"change"`
	}

	if err := cur.All(context.Background(), &whitewoodPlanData); err != nil {
		log.Println(err)
	}
	for i := 0; i < len(whitewoodPlanData); i++ {
		if len(whitewoodPlanData[i].Plans) >= 2 && whitewoodPlanData[i].Plans[1] != 0 {
			whitewoodPlanData[i].Change = whitewoodPlanData[i].Plans[1] - whitewoodPlanData[i].Plan
		} else {
			whitewoodPlanData[i].Change = 0
		}
	}

	// get inventory
	cur, err = s.mgdb.Collection("whitewood").Find(context.Background(), bson.M{"type": "Inventory"}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(2))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var whitewoodInventoryData []struct {
		Prodtype     string    `bson:"prodtype" json:"prodtype"`
		Inventory    float64   `bson:"inventory" json:"inventory"`
		CreatedAt    time.Time `bson:"createdat" json:"createdat"`
		CreatedAtStr string    `json:"createdatstr"`
	}

	if err := cur.All(context.Background(), &whitewoodInventoryData); err != nil {
		log.Println(err)
	}

	for i := 0; i < len(whitewoodInventoryData); i++ {
		whitewoodInventoryData[i].CreatedAtStr = whitewoodInventoryData[i].CreatedAt.Add(7 * time.Hour).Format("15h04 date 2/1")
	}
	// get target
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"name": "whitewood total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -15))}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var whitewoodTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &whitewoodTarget); err != nil {
		log.Println(err)
	}

	// get time of latest update
	sr := s.mgdb.Collection("whitewood").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var LastReport struct {
		Createdat time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&LastReport); err != nil {
		log.Println(err)
	}
	whitewoodUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

	template.Must(template.ParseFiles("templates/pages/dashboard/whitewood.html")).Execute(w, map[string]interface{}{
		"whitewoodData":          whitewoodData,
		"whitewoodPlanData":      whitewoodPlanData,
		"avgdata":                avgdata,
		"whitewoodInventoryData": whitewoodInventoryData,
		"whitewoodTarget":        whitewoodTarget,
		"whitewoodUpTime":        whitewoodUpTime,
	})
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/loadwoodfinish - load woodfinish area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) d_loadwoodfinish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.factory", 1}, {"_id.prodtype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var woodfinishData []struct {
		Date  string  `bson:"date" json:"date"`
		Type  string  `bson:"type" json:"type"`
		Value float64 `bson:"value" json:"value"`
	}
	if err := cur.All(context.Background(), &woodfinishData); err != nil {
		log.Println(err)
	}

	// get plan data
	cur, err = s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
		// {{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}}}}},
		{{"$sort", bson.M{"createdat": -1}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "plantype": "$plantype"}, "plan": bson.M{"$first": "$plan"}, "plans": bson.M{"$firstN": bson.M{"input": "$plan", "n": 2}}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.plantype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "plantype": "$_id.plantype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var woodfinishPlanData []struct {
		Date     string    `bson:"date" json:"date"`
		Plantype string    `bson:"plantype" json:"plantype"`
		Plan     float64   `bson:"plan" json:"plan"`
		Plans    []float64 `bson:"plans" json:"plans"`
		Change   float64   `json:"change"`
	}

	if err := cur.All(context.Background(), &woodfinishPlanData); err != nil {
		log.Println(err)
	}
	for i := 0; i < len(woodfinishPlanData); i++ {
		if len(woodfinishPlanData[i].Plans) >= 2 && woodfinishPlanData[i].Plans[1] != 0 {
			woodfinishPlanData[i].Change = woodfinishPlanData[i].Plans[1] - woodfinishPlanData[i].Plan
		} else {
			woodfinishPlanData[i].Change = 0
		}
	}

	// get inventory
	cur, err = s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "Inventory", "factory": bson.M{"$exists": true}}}},
		{{"$sort", bson.M{"createdat": -1}}},
		{{"$limit", 20}},
		{{"$group", bson.M{"_id": bson.M{"factory": "$factory", "prodtype": "$prodtype"}, "inventory": bson.M{"$first": "$inventory"}, "createdat": bson.M{"$first": "$createdat"}}}},
		{{"$set", bson.M{"type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
		{{"$sort", bson.M{"type": 1}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var woodfinishInventoryData []struct {
		Type         string    `bson:"type" json:"type"`
		Inventory    float64   `bson:"inventory" json:"inventory"`
		CreatedAt    time.Time `bson:"createdat"`
		CreatedAtStr string    `json:"createdat"`
	}
	if err := cur.All(context.Background(), &woodfinishInventoryData); err != nil {
		log.Println(err)
	}

	for i := 0; i < len(woodfinishInventoryData); i++ {
		woodfinishInventoryData[i].CreatedAtStr = woodfinishInventoryData[i].CreatedAt.Add(7 * time.Hour).Format("15h04 date 2/1")
	}

	// get target
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		// {{"$match", bson.M{"name": "woodfinish total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$match", bson.M{"name": "woodfinish total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var woodfinishTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &woodfinishTarget); err != nil {
		log.Println(err)
	}

	// get time of latest update
	sr := s.mgdb.Collection("woodfinish").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var LastReport struct {
		Createdat time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&LastReport); err != nil {
		log.Println(err)
	}
	woodfinishUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

	template.Must(template.ParseFiles("templates/pages/dashboard/woodfinish.html")).Execute(w, map[string]interface{}{
		"woodfinishData":          woodfinishData,
		"woodfinishPlanData":      woodfinishPlanData,
		"woodfinishInventoryData": woodfinishInventoryData,
		"woodfinishTarget":        woodfinishTarget,
		"woodfinishUpTime":        woodfinishUpTime,
	})
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/loadfinemill - load finemill area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) d_loadfinemill(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/loadpack - load pack area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) d_loadpack(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("pack").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.factory", 1}, {"_id.prodtype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	var packData []struct {
		Date  string  `bson:"date" json:"date"`
		Type  string  `bson:"type" json:"type"`
		Value float64 `bson:"value" json:"value"`
	}
	if err := cur.All(context.Background(), &packData); err != nil {
		log.Println(err)
	}

	// get plan data
	cur, err = s.mgdb.Collection("pack").Aggregate(context.Background(), mongo.Pipeline{
		// {{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}}}}},
		{{"$sort", bson.M{"createdat": -1}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "plantype": "$plantype"}, "plan": bson.M{"$first": "$plan"}, "plans": bson.M{"$firstN": bson.M{"input": "$plan", "n": 2}}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.plantype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "plantype": "$_id.plantype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var packPlanData []struct {
		Date     string    `bson:"date" json:"date"`
		Plantype string    `bson:"plantype" json:"plantype"`
		Plan     float64   `bson:"plan" json:"plan"`
		Plans    []float64 `bson:"plans" json:"plans"`
		Change   float64   `json:"change"`
	}

	if err := cur.All(context.Background(), &packPlanData); err != nil {
		log.Println(err)
	}
	for i := 0; i < len(packPlanData); i++ {
		if len(packPlanData[i].Plans) >= 2 && packPlanData[i].Plans[1] != 0 {
			packPlanData[i].Change = packPlanData[i].Plans[1] - packPlanData[i].Plan
		} else {
			packPlanData[i].Change = 0
		}
	}

	// get inventory
	cur, err = s.mgdb.Collection("pack").Find(context.Background(), bson.M{"type": "Inventory"}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(2))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var packInventoryData []struct {
		Prodtype     string    `bson:"prodtype" json:"prodtype"`
		Inventory    float64   `bson:"inventory" json:"inventory"`
		CreatedAt    time.Time `bson:"createdat" json:"createdat"`
		CreatedAtStr string    `json:"createdatstr"`
	}

	if err := cur.All(context.Background(), &packInventoryData); err != nil {
		log.Println(err)
	}

	for i := 0; i < len(packInventoryData); i++ {
		packInventoryData[i].CreatedAtStr = packInventoryData[i].CreatedAt.Add(7 * time.Hour).Format("15h04 date 2/1")
	}

	// get target
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		// {{"$match", bson.M{"name": "packing total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$match", bson.M{"name": "packing total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var packTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &packTarget); err != nil {
		log.Println(err)
	}
	// get time of latest update
	sr := s.mgdb.Collection("pack").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var LastReport struct {
		Createdat time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&LastReport); err != nil {
		log.Println(err)
	}
	packUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

	template.Must(template.ParseFiles("templates/pages/dashboard/pack.html")).Execute(w, map[string]interface{}{
		"packData":          packData,
		"packPlanData":      packPlanData,
		"packInventoryData": packInventoryData,
		"packTarget":        packTarget,
		"packUpTime":        packUpTime,
	})
}

// router.GET("/dashboard/loadslicing", s.d_loadslicing)
func (s *Server) d_loadslicing(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// cur, err := s.mgdb.Collection("slicing").Aggregate(context.Background(), mongo.Pipeline{
	// 	{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
	// 	{{"$group", bson.M{"_id": bson.M{"date": "$date", "prodtype": "$prodtype"}, "qty": bson.M{"$sum": "$qty"}}}},
	// 	{{"$sort", bson.D{{"_id.date", 1}, {"_id.prodtype", 1}}}},
	// 	{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "prodtype": "$_id.prodtype"}}},
	// 	{{"$unset", "_id"}},
	// })
	// if err != nil {
	// 	log.Println(err)
	// }
	// var slicingData []struct {
	// 	Date     string  `bson:"date" json:"date"`
	// 	Prodtype string  `bson:"prodtype" json:"prodtype"`
	// 	Qty      float64 `bson:"qty" json:"qty"`
	// }
	// if err := cur.All(context.Background(), &slicingData); err != nil {
	// 	log.Println(err)
	// }
	// // get target of slicing
	// cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
	// 	{{"$match", bson.M{"name": "slicing total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
	// 	{{"$sort", bson.M{"date": 1}}},
	// 	{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	// })
	// if err != nil {
	// 	log.Println(err)
	// }
	// var slicingTarget []struct {
	// 	Date  string  `bson:"date" json:"date"`
	// 	Value float64 `bson:"value" json:"value"`
	// }
	// if err = cur.All(context.Background(), &slicingTarget); err != nil {
	// 	log.Println(err)
	// }
	// // get last update time of slicing
	// slicingSr := s.mgdb.Collection("slicing").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	// if slicingSr.Err() != nil {
	// 	log.Println(slicingSr.Err())
	// }
	// var slicingLastReport struct {
	// 	CreatedDate time.Time `bson:"createdat" json:"createdat"`
	// }
	// if err := slicingSr.Decode(&slicingLastReport); err != nil {
	// 	log.Println(err)
	// }

	// slicingUpTime := slicingLastReport.CreatedDate.Add(7 * time.Hour).Format("15:04")

	// template.Must(template.ParseFiles("templates/pages/dashboard/slicing.html")).Execute(w, map[string]interface{}{
	// 	"slicingData":   slicingData,
	// 	"slicingTarget": slicingTarget,
	// 	"slicingUpTime": slicingUpTime,
	// })

	cur, err := s.mgdb.Collection("output").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"section": "1.Slice"}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -25))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "type": "$type"}, "qty": bson.M{"$sum": "$qty"}}}},
		{{"$sort", bson.M{"_id.date": 1, "_id.type": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "prodtype": "$_id.type"}}},
		{{"$unset", bson.A{"_id.date", "_id.type"}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var slicingData []struct {
		Date     string  `bson:"date" json:"date"`
		Prodtype string  `bson:"prodtype" json:"prodtype"`
		Qty      float64 `bson:"qty" json:"qty"`
	}
	if err := cur.All(context.Background(), &slicingData); err != nil {
		log.Println(err)
	}

	// get target
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		// {{"$match", bson.M{"name": "packing total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$match", bson.M{"name": "slicing total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -25))}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var slicingTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &slicingTarget); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/dashboard/slicing.html")).Execute(w, map[string]interface{}{
		"slicingData":   slicingData,
		"slicingTarget": slicingTarget,
		// "slicingUpTime": slicingUpTime,
	})
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/loadwoodrecovery - load woodrecovery area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) d_loadwoodrecovery(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("woodrecovery").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$sort", bson.D{{"date", 1}, {"prodtype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var woodrecoveryData []struct {
		Date     string  `bson:"date" json:"date"`
		Prodtype string  `bson:"prodtype" json:"prodtype"`
		Rate     float64 `bson:"rate" json:"rate"`
	}
	if err := cur.All(context.Background(), &woodrecoveryData); err != nil {
		log.Println(err)
	}
	// get last update time
	sr := s.mgdb.Collection("woodrecovery").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var LastReport struct {
		Createdat time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&LastReport); err != nil {
		log.Println(err)
	}
	woodrecoveryUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")
	template.Must(template.ParseFiles("templates/pages/dashboard/woodrecovery.html")).Execute(w, map[string]interface{}{
		"woodrecoveryData":   woodrecoveryData,
		"woodrecoveryUpTime": woodrecoveryUpTime,
	})
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/loadquality - load quality area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) d_loadquality(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("quality").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": time.Now().AddDate(0, 0, -13).Format("2006-01-02")}}, bson.M{"date": bson.M{"$lte": time.Now().Format("2006-01-02")}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "section": "$section"}, "checkedqty": bson.M{"$sum": "$checkedqty"}, "failedqty": bson.M{"$sum": "$failedqty"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.section", 1}}}},
		{{"$set", bson.M{"date": "$_id.date", "section": "$_id.section"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var qualityChartData []struct {
		Date       string `bson:"date" json:"date"`
		Section    string `bson:"section" json:"section"`
		CheckedQty int    `bson:"checkedqty" json:"checkedqty"`
		FailedQty  int    `bson:"failedqty" json:"failedqty"`
	}
	if err := cur.All(context.Background(), &qualityChartData); err != nil {
		log.Println(err)
	}
	for i := 0; i < len(qualityChartData); i++ {
		tmp, _ := time.Parse("2006-01-02", qualityChartData[i].Date)
		qualityChartData[i].Date = tmp.Format("02 Jan")
	}
	template.Must(template.ParseFiles("templates/pages/dashboard/quality.html")).Execute(w, map[string]interface{}{
		"qualityChartData": qualityChartData,
	})
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/loaddowntime - load downtime area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) d_loaddowntime(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/loadsixs - load 6S area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) d_loadsixs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/loadsafety
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) d_loadsafety(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/panelcnc/getchart - change chart of panelcnc area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dpc_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("panelcnccharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("panelcncFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("panelcncToDate"))

	switch pickedChart {

	case "machinechart":
		pipeline := mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate.AddDate(0, 0, 1))}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "machine": "$machine"}, "qty": bson.M{"$sum": "$qty"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.machine", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "machine": "$_id.machine"}}},
			{{"$unset", "_id"}},
		}
		cur, err := s.mgdb.Collection("panelcnc").Aggregate(context.Background(), pipeline)
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var panelChartData []struct {
			Date    string  `bson:"date" json:"date"`
			Machine string  `bson:"machine" json:"machine"`
			Qty     float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &panelChartData); err != nil {
			log.Println(err)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/panelcnc_machinechart.html")).Execute(w, map[string]interface{}{
			"panelChartData": panelChartData,
		})

	case "general":
		pipeline := mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "qty": bson.M{"$sum": "$qty"}}}},
			{{"$sort", bson.D{{"_id", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
			{{"$unset", "_id"}},
		}
		cur, err := s.mgdb.Collection("panelcnc").Aggregate(context.Background(), pipeline)
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var panelChartData []struct {
			Date string  `bson:"date" json:"date"`
			Qty  float64 `bson:"qty" json:"qty"`
		}

		if err := cur.All(context.Background(), &panelChartData); err != nil {
			log.Println(err)
		}
		// get target of panelcnc
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"name": "panelcnc total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var panelcncTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &panelcncTarget); err != nil {
			log.Println(err)
		}

		// get time of latest update
		sr := s.mgdb.Collection("panelcnc").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
		if sr.Err() != nil {
			log.Println(sr.Err())
		}
		var LastReport struct {
			Createdat time.Time `bson:"createdat" json:"createdat"`
		}
		if err := sr.Decode(&LastReport); err != nil {
			log.Println(err)
		}
		panelcncUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

		template.Must(template.ParseFiles("templates/pages/dashboard/panelcnc_totalchart.html")).Execute(w, map[string]interface{}{
			"panelChartData": panelChartData,
			"panelcncTarget": panelcncTarget,
			"panelcncUpTime": panelcncUpTime,
		})

	case "efficiency":
		cur, err := s.mgdb.Collection("panelcnc").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "qty": bson.M{"$sum": "$qty"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var panelcncData []struct {
			Date string  `bson:"date" json:"date"`
			Qty  float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &panelcncData); err != nil {
			log.Println(err)
		}
		// get workhr of pannelcnc from manhr
		cur, err = s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"section": "panelcnc", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println("err")
		}
		var panelcncManhr []struct {
			Date   string  `bson:"date" json:"date"`
			HC     int     `bson:"hc" json:"hc"`
			Workhr float64 `bson:"workhr" json:"workhr"`
		}
		if err = cur.All(context.Background(), &panelcncManhr); err != nil {
			log.Println("err")
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/panelcnc_efficiencychart.html")).Execute(w, map[string]interface{}{
			"panelcncData":  panelcncData,
			"panelcncManhr": panelcncManhr,
		})
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/assembly/getchart - change chart of assembly area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) da_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("assemblycharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("assemblyFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("assemblyToDate"))

	switch pickedChart {
	case "general":
		cur, err := s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.factory", 1}, {"_id.prodtype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var assemblyData []struct {
			Date  string  `bson:"date" json:"date"`
			Type  string  `bson:"type" json:"type"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &assemblyData); err != nil {
			log.Println(err)
		}

		// get plan data
		cur, err = s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
			// {{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -12))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
			{{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}}}}},
			{{"$sort", bson.M{"createdat": -1}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "plantype": "$plantype"}, "plan": bson.M{"$first": "$plan"}, "plans": bson.M{"$firstN": bson.M{"input": "$plan", "n": 2}}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.plantype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "plantype": "$_id.plantype"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var assemblyPlanData []struct {
			Date     string    `bson:"date" json:"date"`
			Plantype string    `bson:"plantype" json:"plantype"`
			Plan     float64   `bson:"plan" json:"plan"`
			Plans    []float64 `bson:"plans" json:"plans"`
			Change   float64   `json:"change"`
		}

		if err := cur.All(context.Background(), &assemblyPlanData); err != nil {
			log.Println(err)
		}
		for i := 0; i < len(assemblyPlanData); i++ {
			if len(assemblyPlanData[i].Plans) >= 2 && assemblyPlanData[i].Plans[1] != 0 {
				assemblyPlanData[i].Change = assemblyPlanData[i].Plans[1] - assemblyPlanData[i].Plan
			} else {
				assemblyPlanData[i].Change = 0
			}
		}

		// get inventory
		cur, err = s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"type": "Inventory", "factory": bson.M{"$exists": true}}}},
			{{"$sort", bson.M{"createdat": -1}}},
			{{"$limit", 20}},
			{{"$group", bson.M{"_id": bson.M{"factory": "$factory", "prodtype": "$prodtype"}, "inventory": bson.M{"$first": "$inventory"}, "createdat": bson.M{"$first": "$createdat"}}}},
			{{"$set", bson.M{"type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
			{{"$sort", bson.M{"type": 1}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var assemblyInventoryData []struct {
			Type         string    `bson:"type" json:"type"`
			Inventory    float64   `bson:"inventory" json:"inventory"`
			CreatedAt    time.Time `bson:"createdat"`
			CreatedAtStr string    `json:"createdat"`
		}
		if err := cur.All(context.Background(), &assemblyInventoryData); err != nil {
			log.Println(err)
		}

		for i := 0; i < len(assemblyInventoryData); i++ {
			assemblyInventoryData[i].CreatedAtStr = assemblyInventoryData[i].CreatedAt.Add(7 * time.Hour).Format("15h04 date 2/1")
		}

		// get target
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			// {{"$match", bson.M{"name": "assembly total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
			{{"$match", bson.M{"name": "assembly total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var assemblyTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &assemblyTarget); err != nil {
			log.Println(err)
		}

		// get time of latest update
		sr := s.mgdb.Collection("assembly").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
		if sr.Err() != nil {
			log.Println(sr.Err())
		}
		var LastReport struct {
			Createdat time.Time `bson:"createdat" json:"createdat"`
		}
		if err := sr.Decode(&LastReport); err != nil {
			log.Println(err)
		}
		assemblyUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")
		template.Must(template.ParseFiles("templates/pages/dashboard/assembly_generalchart.html")).Execute(w, map[string]interface{}{
			"assemblyData":          assemblyData,
			"assemblyPlanData":      assemblyPlanData,
			"assemblyInventoryData": assemblyInventoryData,
			"assemblyTarget":        assemblyTarget,
			"assemblyUpTime":        assemblyUpTime,
		})

	case "detail":
		cur, err := s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
			// {{"$match", bson.M{"itemtype": "whole"}}},
			{{"$match", bson.M{"$and": bson.A{bson.M{"itemtype": "whole"}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
			{{"$sort", bson.D{{"type", 1}, {"date", 1}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var assemblyChartData []struct {
			Date  string  `bson:"date" json:"date"`
			Type  string  `bson:"type" json:"type"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &assemblyChartData); err != nil {
			log.Println(err)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/assembly_detailchart.html")).Execute(w, map[string]interface{}{
			"assemblyChartData": assemblyChartData,
		})

	case "value-target":
		cur, err := s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.factory", 1}, {"_id.prodtype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var assemblyData []struct {
			Date  string  `bson:"date" json:"date"`
			Type  string  `bson:"type" json:"type"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &assemblyData); err != nil {
			log.Println(err)
		}

		// get target
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"name": "assembly total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -15))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var assemblyTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &assemblyTarget); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/assembly_valuetargetchart.html")).Execute(w, map[string]interface{}{
			"assemblyData":   assemblyData,
			"assemblyTarget": assemblyTarget,
		})

	case "efficiency":
		cur, err := s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var assemblyData []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &assemblyData); err != nil {
			log.Println(err)
		}
		//get manhr of assembly
		cur, err = s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"section": "assembly", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var assemblyManhr []struct {
			Date   string  `bson:"date" json:"date"`
			HC     int     `bson:"hc" json:"hc"`
			Workhr float64 `bson:"workhr" json:"workhr"`
		}
		if err = cur.All(context.Background(), &assemblyManhr); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/assembly_efficiencychart.html")).Execute(w, map[string]interface{}{
			"assemblyData":  assemblyData,
			"assemblyManhr": assemblyManhr,
		})
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/woodfinish/getchart - change chart of woodfinish area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dw_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("woodfinishcharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("woodfinishFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("woodfinishToDate"))

	switch pickedChart {
	case "general":
		cur, err := s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "itemtype": "$itemtype"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.itemtype", -1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": "$_id.itemtype"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var woodfinishChartData []struct {
			Date  string  `bson:"date" json:"date"`
			Type  string  `bson:"type" json:"type"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &woodfinishChartData); err != nil {
			log.Println(err)
		}
		// get target of wood finish
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"name": "woodfinish total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var woodfinishTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &woodfinishTarget); err != nil {
			log.Println(err)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/wf_generalchart.html")).Execute(w, map[string]interface{}{
			"woodfinishChartData": woodfinishChartData,
			"woodfinishTarget":    woodfinishTarget,
		})

	case "detail":
		cur, err := s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"itemtype": "whole"}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
			{{"$sort", bson.D{{"type", 1}, {"date", 1}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var woodfinishChartData []struct {
			Date  string  `bson:"date" json:"date"`
			Type  string  `bson:"type" json:"type"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &woodfinishChartData); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/wf_detailchart.html")).Execute(w, map[string]interface{}{
			"woodfinishChartData": woodfinishChartData,
		})

	case "value-target":
		cur, err := s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.factory", 1}, {"_id.prodtype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var woodfinishData []struct {
			Date  string  `bson:"date" json:"date"`
			Type  string  `bson:"type" json:"type"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &woodfinishData); err != nil {
			log.Println(err)
		}

		// get plan data
		cur, err = s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
			// {{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
			{{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}}}}},
			{{"$sort", bson.M{"createdat": -1}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "plantype": "$plantype"}, "plan": bson.M{"$first": "$plan"}, "plans": bson.M{"$firstN": bson.M{"input": "$plan", "n": 2}}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.plantype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "plantype": "$_id.plantype"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var woodfinishPlanData []struct {
			Date     string    `bson:"date" json:"date"`
			Plantype string    `bson:"plantype" json:"plantype"`
			Plan     float64   `bson:"plan" json:"plan"`
			Plans    []float64 `bson:"plans" json:"plans"`
			Change   float64   `json:"change"`
		}

		if err := cur.All(context.Background(), &woodfinishPlanData); err != nil {
			log.Println(err)
		}
		for i := 0; i < len(woodfinishPlanData); i++ {
			if len(woodfinishPlanData[i].Plans) >= 2 && woodfinishPlanData[i].Plans[1] != 0 {
				woodfinishPlanData[i].Change = woodfinishPlanData[i].Plans[1] - woodfinishPlanData[i].Plan
			} else {
				woodfinishPlanData[i].Change = 0
			}
		}

		// get inventory
		cur, err = s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"type": "Inventory", "factory": bson.M{"$exists": true}}}},
			{{"$sort", bson.M{"createdat": -1}}},
			{{"$limit", 20}},
			{{"$group", bson.M{"_id": bson.M{"factory": "$factory", "prodtype": "$prodtype"}, "inventory": bson.M{"$first": "$inventory"}, "createdat": bson.M{"$first": "$createdat"}}}},
			{{"$set", bson.M{"type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
			{{"$sort", bson.M{"type": 1}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var woodfinishInventoryData []struct {
			Type         string    `bson:"type" json:"type"`
			Inventory    float64   `bson:"inventory" json:"inventory"`
			CreatedAt    time.Time `bson:"createdat"`
			CreatedAtStr string    `json:"createdat"`
		}
		if err := cur.All(context.Background(), &woodfinishInventoryData); err != nil {
			log.Println(err)
		}

		for i := 0; i < len(woodfinishInventoryData); i++ {
			woodfinishInventoryData[i].CreatedAtStr = woodfinishInventoryData[i].CreatedAt.Add(7 * time.Hour).Format("15h04 date 2/1")
		}

		// get target
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			// {{"$match", bson.M{"name": "woodfinish total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
			{{"$match", bson.M{"name": "woodfinish total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var woodfinishTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &woodfinishTarget); err != nil {
			log.Println(err)
		}

		// get time of latest update
		sr := s.mgdb.Collection("woodfinish").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
		if sr.Err() != nil {
			log.Println(sr.Err())
		}
		var LastReport struct {
			Createdat time.Time `bson:"createdat" json:"createdat"`
		}
		if err := sr.Decode(&LastReport); err != nil {
			log.Println(err)
		}
		woodfinishUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

		template.Must(template.ParseFiles("templates/pages/dashboard/wf_valuetargetchart.html")).Execute(w, map[string]interface{}{
			"woodfinishData":          woodfinishData,
			"woodfinishPlanData":      woodfinishPlanData,
			"woodfinishInventoryData": woodfinishInventoryData,
			"woodfinishTarget":        woodfinishTarget,
			"woodfinishUpTime":        woodfinishUpTime,
		})

	case "efficiency":
		cur, err := s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var woodfinishData []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &woodfinishData); err != nil {
			log.Println(err)
		}
		//get manhr of woodfinish
		cur, err = s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"section": "woodfinish", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var woodfinishManhr []struct {
			Date   string  `bson:"date" json:"date"`
			HC     int     `bson:"hc" json:"hc"`
			Workhr float64 `bson:"workhr" json:"workhr"`
		}
		if err = cur.All(context.Background(), &woodfinishManhr); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/wf_efficiencychart.html")).Execute(w, map[string]interface{}{
			"woodfinishData":  woodfinishData,
			"woodfinishManhr": woodfinishManhr,
		})
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// router.POST("/dashboard/whitewood/getchart", s.dww_getchart)
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dww_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("whitewoodcharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("whitewoodFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("whitewoodToDate"))

	switch pickedChart {
	case "value-target":
		cur, err := s.mgdb.Collection("whitewood").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.prodtype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": "$_id.prodtype"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var whitewoodData []struct {
			Date  string  `bson:"date" json:"date"`
			Type  string  `bson:"type" json:"type"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &whitewoodData); err != nil {
			log.Println(err)
		}

		// get plan data
		cur, err = s.mgdb.Collection("whitewood").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}}}}},
			{{"$sort", bson.M{"createdat": -1}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "plantype": "$plantype"}, "plan": bson.M{"$first": "$plan"}, "plans": bson.M{"$firstN": bson.M{"input": "$plan", "n": 2}}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.plantype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "plantype": "$_id.plantype"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var whitewoodPlanData []struct {
			Date     string    `bson:"date" json:"date"`
			Plantype string    `bson:"plantype" json:"plantype"`
			Plan     float64   `bson:"plan" json:"plan"`
			Plans    []float64 `bson:"plans" json:"plans"`
			Change   float64   `json:"change"`
		}

		if err := cur.All(context.Background(), &whitewoodPlanData); err != nil {
			log.Println(err)
		}
		for i := 0; i < len(whitewoodPlanData); i++ {
			if len(whitewoodPlanData[i].Plans) >= 2 && whitewoodPlanData[i].Plans[1] != 0 {
				whitewoodPlanData[i].Change = whitewoodPlanData[i].Plans[1] - whitewoodPlanData[i].Plan
			} else {
				whitewoodPlanData[i].Change = 0
			}
		}

		// get inventory
		cur, err = s.mgdb.Collection("whitewood").Find(context.Background(), bson.M{"type": "Inventory"}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(2))
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var whitewoodInventoryData []struct {
			Prodtype     string    `bson:"prodtype" json:"prodtype"`
			Inventory    float64   `bson:"inventory" json:"inventory"`
			CreatedAt    time.Time `bson:"createdat" json:"createdat"`
			CreatedAtStr string    `json:"createdatstr"`
		}

		if err := cur.All(context.Background(), &whitewoodInventoryData); err != nil {
			log.Println(err)
		}

		for i := 0; i < len(whitewoodInventoryData); i++ {
			whitewoodInventoryData[i].CreatedAtStr = whitewoodInventoryData[i].CreatedAt.Add(7 * time.Hour).Format("15h04 date 2/1")
		}
		// get target
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"name": "whitewood total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var whitewoodTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &whitewoodTarget); err != nil {
			log.Println(err)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/whitewood_generalchart.html")).Execute(w, map[string]interface{}{
			"whitewoodData":          whitewoodData,
			"whitewoodPlanData":      whitewoodPlanData,
			"whitewoodInventoryData": whitewoodInventoryData,
			"whitewoodTarget":        whitewoodTarget,
		})

	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/cutting/getchart - change chart of cutting area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dc_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("cuttingcharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("cuttingFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("cuttingToDate"))

	switch pickedChart {
	case "general":
		// get data for cutting chart
		pipeline := mongo.Pipeline{
			{{"$match", bson.M{"type": "report", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "is25reeded": "$is25reeded"}, "qty": bson.M{"$sum": "$qtycbm"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.is25reeded", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "is25": "$_id.is25reeded"}}},
			{{"$unset", "_id"}},
		}

		cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), pipeline)
		if err != nil {
			log.Println(err)
			return
		}
		defer cur.Close(context.Background())
		var cuttingData []struct {
			Date string  `bson:"date" json:"date"`
			Is25 bool    `bson:"is25" json:"is25"`
			Qty  float64 `bson:"qty" json:"qty"`
		}

		if err = cur.All(context.Background(), &cuttingData); err != nil {
			log.Println(err)
			return
		}

		//get wood return
		cur, err = s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"type": "return", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "is25": "$is25"}, "qty": bson.M{"$sum": "$qtycbm"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.is25", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "is25": "$_id.is25"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
			return
		}
		var cuttingReturnData []struct {
			Date string  `bson:"date" json:"date"`
			Is25 bool    `bson:"is25" json:"is25"`
			Qty  float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &cuttingReturnData); err != nil {
			log.Println(err)
			return
		}

		//get fine wood
		cur, err = s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"type": "fine", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "is25reeded": "$is25reeded"}, "qty": bson.M{"$sum": "$qtycbm"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.is25reeded", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "is25reeded": "$_id.is25reeded"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
			return
		}
		var cuttingFineData []struct {
			Date       string  `bson:"date" json:"date"`
			Is25reeded bool    `bson:"is25reeded" json:"is25reeded"`
			Qty        float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &cuttingFineData); err != nil {
			log.Println(err)
			return
		}

		//get target data for leftchart
		sr := s.mgdb.Collection("cutting").FindOne(context.Background(), bson.M{"type": "target"}, options.FindOne().SetSort(bson.M{"startdate": -1}))
		if sr.Err() != nil {
			log.Println(sr.Err())
		}
		var targetactualData struct {
			Name      string    `bson:"name" json:"name"`
			StartDate time.Time `bson:"startdate"`
			EnddDate  time.Time `bson:"enddate"`
			Detail    []struct {
				Type   string  `bson:"type" json:"type"`
				Target float64 `bson:"target" json:"target"`
			} `bson:"detail" json:"detail"`
			StartDateStr string `json:"startdate"`
			EndDateStr   string `json:"enddate"`
		}
		if err := sr.Decode(&targetactualData); err != nil {
			log.Println(err)
		}
		targetactualData.StartDateStr = targetactualData.StartDate.Format("02/01/2006")
		targetactualData.EndDateStr = targetactualData.EnddDate.Format("02/01/2006")

		cur, err = s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"type": "fine"}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(targetactualData.StartDate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(targetactualData.EnddDate)}}}}}},
			{{"$set", bson.M{"is25reeded": bson.M{"$ifNull": bson.A{"$is25reeded", false}}}}},
			{{"$group", bson.M{"_id": "$is25reeded", "qty": bson.M{"$sum": "$qtycbm"}}}},
			{{"$sort", bson.D{{"_id", 1}}}},
			{{"$set", bson.M{"prodtype": "$_id"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
			return
		}
		defer cur.Close(context.Background())
		var cuttingProdtypeData []struct {
			Prodtype bool    `bson:"prodtype" json:"prodtype"`
			Qty      float64 `bson:"qty" json:"qty"`
		}

		if err = cur.All(context.Background(), &cuttingProdtypeData); err != nil {
			log.Println(err)
			return
		}

		if len(cuttingProdtypeData) == 1 {
			if cuttingProdtypeData[0].Prodtype {
				cuttingProdtypeData = append(cuttingProdtypeData, struct {
					Prodtype bool    `bson:"prodtype" json:"prodtype"`
					Qty      float64 `bson:"qty" json:"qty"`
				}{
					Prodtype: false, Qty: 0,
				})
			} else {
				cuttingProdtypeData = append(cuttingProdtypeData, struct {
					Prodtype bool    `bson:"prodtype" json:"prodtype"`
					Qty      float64 `bson:"qty" json:"qty"`
				}{
					Prodtype: true, Qty: 0,
				})
			}
		}

		//get target line data of cutting
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"name": "cutting total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var cuttingTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &cuttingTarget); err != nil {
			log.Println(err)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/cutting_generalchart.html")).Execute(w, map[string]interface{}{
			"cuttingData":         cuttingData,
			"cuttingReturnData":   cuttingReturnData,
			"cuttingFineData":     cuttingFineData,
			"targetactualData":    targetactualData,
			"cuttingProdtypeData": cuttingProdtypeData,
			"cuttingTarget":       cuttingTarget,
		})

	case "woodtype":
		pipeline := mongo.Pipeline{
			{{"$match", bson.M{"type": "report", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$woodtype", "qty": bson.M{"$sum": "$qtycbm"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"woodtype": "$_id"}}},
			{{"$unset", "_id"}},
		}
		cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), pipeline, options.Aggregate())
		if err != nil {
			log.Println(err)
			return
		}
		defer cur.Close(context.Background())
		var cuttingData []struct {
			Woodtype string  `bson:"woodtype" json:"woodtype"`
			Qty      float64 `bson:"qty" json:"qty"`
		}
		if err = cur.All(context.Background(), &cuttingData); err != nil {
			log.Println(err)
			return
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/cutting_woodtypechart.html")).Execute(w, map[string]interface{}{
			"cuttingData": cuttingData,
		})

	case "efficiency":
		cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"type": "report", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "prodtype": "$prodtype"}, "qty": bson.M{"$sum": "$qtycbm"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.prodtype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "prodtype": "$_id.prodtype"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var cuttingData []struct {
			Date     string  `bson:"date" json:"date"`
			ProdType string  `bson:"prodtype" json:"prodtype"`
			Qty      float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &cuttingData); err != nil {
			log.Println(err)
		}

		//get manhr of cutting
		cur, err = s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"section": "cutting", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var cuttingManhr []struct {
			Date   string  `bson:"date" json:"date"`
			HC     int     `bson:"hc" json:"hc"`
			WorkHr float64 `bson:"workhr" json:"workhr"`
			Qty    float64
		}
		if err = cur.All(context.Background(), &cuttingManhr); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/cutting_efficiencychart.html")).Execute(w, map[string]interface{}{
			"cuttingData":  cuttingData,
			"cuttingManhr": cuttingManhr,
		})
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/lamination/getchart - change chart of cutting area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dl_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("laminationcharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("laminationFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("laminationToDate"))

	switch pickedChart {
	case "general":
		// get data for lamination
		cur, err := s.mgdb.Collection("lamination").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "prodtype": "$prodtype"}, "qty": bson.M{"$sum": "$qty"}}}},
			{{"$sort", bson.M{"_id.date": 1, "_id.prodtype": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "prodtype": "$_id.prodtype"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		var laminationChartData []struct {
			Date     string  `bson:"date" json:"date"`
			Prodtype string  `bson:"prodtype" json:"prodtype"`
			Qty      float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &laminationChartData); err != nil {
			log.Println(err)
		}
		// get target of lamination
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"name": "lamination total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var laminationTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &laminationTarget); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/lamination_generalchart.html")).Execute(w, map[string]interface{}{
			"laminationChartData": laminationChartData,
			"laminationTarget":    laminationTarget,
		})

	case "efficiency":
		cur, err := s.mgdb.Collection("lamination").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "qty": bson.M{"$sum": "$qty"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var laminationData []struct {
			Date string  `bson:"date" json:"date"`
			Qty  float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &laminationData); err != nil {
			log.Println(err)
		}

		//get manhr of lamination
		cur, err = s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"section": "lamination", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var laminationManhr []struct {
			Date   string  `bson:"date" json:"date"`
			HC     int     `bson:"hc" json:"hc"`
			Workhr float64 `bson:"workhr" json:"workhr"`
		}
		if err = cur.All(context.Background(), &laminationManhr); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/lamination_efficiencychart.html")).Execute(w, map[string]interface{}{
			"laminationData":  laminationData,
			"laminationManhr": laminationManhr,
		})
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/reededline/getchart
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dr_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("reededlinecharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("reededlineFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("reededlineToDate"))

	switch pickedChart {
	case "general":
		cur, err := s.mgdb.Collection("reededline").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "tone": "$tone"}, "qty": bson.M{"$sum": "$qty"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.tone", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "tone": "$_id.tone"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var reededlinedata []struct {
			Date string  `bson:"date" json:"date"`
			Tone string  `bson:"tone" json:"tone"`
			Qty  float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &reededlinedata); err != nil {
			log.Println(err)
		}

		// get data of Gỗ 25 of cutting
		cur, err = s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"is25reeded": true}, bson.M{"type": "report"}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "qty": bson.M{"$sum": "$qtycbm"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		var wood25data []struct {
			Date string  `bson:"date" json:"date"`
			Qty  float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &wood25data); err != nil {
			log.Println(err)
		}

		// get target of reededline
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"name": "reededline total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var reededlineTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &reededlineTarget); err != nil {
			log.Println(err)
		}

		// get time of latest update
		sr := s.mgdb.Collection("reededline").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
		if sr.Err() != nil {
			log.Println(sr.Err())
		}
		var LastReport struct {
			Createdat time.Time `bson:"createdat" json:"createdat"`
		}
		if err := sr.Decode(&LastReport); err != nil {
			log.Println(err)
		}
		reededlineUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

		template.Must(template.ParseFiles("templates/pages/dashboard/reededline_generalchart.html")).Execute(w, map[string]interface{}{
			"reededlinedata":   reededlinedata,
			"wood25data":       wood25data,
			"reededlineTarget": reededlineTarget,
			"reededlineUpTime": reededlineUpTime,
		})

	case "efficiency":
		cur, err := s.mgdb.Collection("reededline").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "qty": bson.M{"$sum": "$qty"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var reededlineData []struct {
			Date string  `bson:"date" json:"date"`
			Qty  float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &reededlineData); err != nil {
			log.Println(err)
		}

		//get manhr of reededline
		cur, err = s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"section": "reededline", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var reededlineManhr []struct {
			Date   string  `bson:"date" json:"date"`
			HC     int     `bson:"hc" json:"hc"`
			Workhr float64 `bson:"workhr" json:"workhr"`
		}
		if err = cur.All(context.Background(), &reededlineManhr); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/reededline_efficiencychart.html")).Execute(w, map[string]interface{}{
			"reededlineData":  reededlineData,
			"reededlineManhr": reededlineManhr,
		})
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/output/getchart
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) do_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("outputcharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("outputFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("outputToDate"))

	switch pickedChart {
	case "reeded":
		cur, err := s.mgdb.Collection("output").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"type": "reeded", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "section": "$section"}, "qty": bson.M{"$sum": "$qty"}, "type": bson.M{"$first": "$type"}}}},
			{{"$set", bson.M{"section": "$_id.section", "date": "$_id.date"}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$group", bson.M{"_id": "$section", "type": bson.M{"$first": "$type"}, "qty": bson.M{"$sum": "$qty"}, "avg": bson.M{"$avg": "$qty"}, "lastdate": bson.M{"$last": "$date"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"section": bson.M{"$substr": bson.A{"$_id", 2, -1}}, "lastdate": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$lastdate"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var reededoutputData []struct {
			Section  string  `bson:"section" json:"section"`
			Type     string  `bson:"type" json:"type"`
			Qty      float64 `bson:"qty" json:"qty"`
			Avg      float64 `bson:"avg" json:"avg"`
			LastDate string  `bson:"lastdate" json:"lastdate"`
		}
		if err := cur.All(context.Background(), &reededoutputData); err != nil {
			log.Println(err)
		}
		// get latest inventory
		sr := s.mgdb.Collection("output").FindOne(context.Background(), bson.M{"date": primitive.NewDateTimeFromTime(todate), "section": "a.Inventory"})
		if sr.Err() != nil {
			log.Println(sr.Err())
		}
		var latestInventory struct {
			Date    time.Time `bson:"date"`
			Section string    `json:"section"`
			Qty     float64   `bson:"qty" json:"qty"`
			DateStr string    `json:"date"`
		}
		if err := sr.Decode(&latestInventory); err != nil {
			log.Println(err)
		}
		latestInventory.Section = "Inventory"
		latestInventory.DateStr = latestInventory.Date.Format("02-01-2006")
		template.Must(template.ParseFiles("templates/pages/dashboard/reededoutput_totalchart.html")).Execute(w, map[string]interface{}{
			"reededoutputData": reededoutputData,
			"latestInventory":  latestInventory,
		})

	case "fir":
		cur, err := s.mgdb.Collection("output").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"type": "fir", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$group", bson.M{"_id": "$section", "type": bson.M{"$first": "$type"}, "qty": bson.M{"$sum": "$qty"}, "avg": bson.M{"$avg": "$qty"}, "lastdate": bson.M{"$last": "$date"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"section": bson.M{"$substr": bson.A{"$_id", 2, -1}}, "lastdate": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$lastdate"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var firoutputData []struct {
			Section  string  `bson:"section" json:"section"`
			Type     string  `bson:"type" json:"type"`
			Qty      float64 `bson:"qty" json:"qty"`
			Avg      float64 `bson:"avg" json:"avg"`
			LastDate string  `bson:"lastdate" json:"lastdate"`
		}
		if err := cur.All(context.Background(), &firoutputData); err != nil {
			log.Println(err)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/firoutput_totalchart.html")).Execute(w, map[string]interface{}{
			"firoutputData": firoutputData,
		})
	}

}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/lamination/getchart - change chart of cutting area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dv_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("veneercharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("veneerFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("veneerToDate"))

	switch pickedChart {
	case "general":
		cur, err := s.mgdb.Collection("veneer").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "type": "$type"}, "qty": bson.M{"$sum": "$qty"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.type", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": "$_id.type"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var veneerChartData []struct {
			Date string  `bson:"date" json:"date"`
			Type string  `bson:"type" json:"type"`
			Qty  float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &veneerChartData); err != nil {
			log.Println(err)
		}
		// get target for veneer
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"name": "veneer total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -20))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var veneerTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &veneerTarget); err != nil {
			log.Println(err)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/veneer_generalchart.html")).Execute(w, map[string]interface{}{
			"veneerChartData": veneerChartData,
			"veneerTarget":    veneerTarget,
		})

	case "efficiency":
		cur, err := s.mgdb.Collection("veneer").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "qty": bson.M{"$sum": "$qty"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var veneerData []struct {
			Date string  `bson:"date" json:"date"`
			Qty  float64 `bson:"qty" json:"qty"`
		}
		if err := cur.All(context.Background(), &veneerData); err != nil {
			log.Println(err)
		}

		//get manhr of veneer
		cur, err = s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"section": "veneer", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var veneerManhr []struct {
			Date   string  `bson:"date" json:"date"`
			HC     int     `bson:"hc" json:"hc"`
			Workhr float64 `bson:"workhr" json:"workhr"`
		}
		if err = cur.All(context.Background(), &veneerManhr); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/veneer_efficiencychart.html")).Execute(w, map[string]interface{}{
			"veneerData":  veneerData,
			"veneerManhr": veneerManhr,
		})
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/finemill/getchart - change chart of finemill area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) df_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/pack/getchart - change chart of pack area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dp_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("packcharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("packFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("packToDate"))

	switch pickedChart {
	case "general":
		cur, err := s.mgdb.Collection("pack").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "itemtype": "$itemtype"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.itemtype", -1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": "$_id.itemtype"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var packChartData []struct {
			Date  string  `bson:"date" json:"date"`
			Type  string  `bson:"type" json:"type"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &packChartData); err != nil {
			log.Println(err)
		}
		// get target for pack
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"name": "packing total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var packingTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &packingTarget); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/pack_generalchart.html")).Execute(w, map[string]interface{}{
			"packChartData": packChartData,
			"packingTarget": packingTarget,
		})

	case "detail":
		cur, err := s.mgdb.Collection("pack").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"itemtype": "whole", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
			{{"$sort", bson.D{{"type", 1}, {"date", 1}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var packChartData []struct {
			Date  string  `bson:"date" json:"date"`
			Type  string  `bson:"type" json:"type"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &packChartData); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/pack_detailchart.html")).Execute(w, map[string]interface{}{
			"packChartData": packChartData,
		})

	case "valuetarget":
		cur, err := s.mgdb.Collection("pack").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "factory": "$factory", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.factory", 1}, {"_id.prodtype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": bson.M{"$concat": bson.A{"X", "$_id.factory", "-", "$_id.prodtype"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		var packData []struct {
			Date  string  `bson:"date" json:"date"`
			Type  string  `bson:"type" json:"type"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &packData); err != nil {
			log.Println(err)
		}

		// get plan data
		cur, err = s.mgdb.Collection("pack").Aggregate(context.Background(), mongo.Pipeline{
			// {{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
			{{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}}}}},
			{{"$sort", bson.M{"createdat": -1}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "plantype": "$plantype"}, "plan": bson.M{"$first": "$plan"}, "plans": bson.M{"$firstN": bson.M{"input": "$plan", "n": 2}}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.plantype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "plantype": "$_id.plantype"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var packPlanData []struct {
			Date     string    `bson:"date" json:"date"`
			Plantype string    `bson:"plantype" json:"plantype"`
			Plan     float64   `bson:"plan" json:"plan"`
			Plans    []float64 `bson:"plans" json:"plans"`
			Change   float64   `json:"change"`
		}

		if err := cur.All(context.Background(), &packPlanData); err != nil {
			log.Println(err)
		}
		for i := 0; i < len(packPlanData); i++ {
			if len(packPlanData[i].Plans) >= 2 && packPlanData[i].Plans[1] != 0 {
				packPlanData[i].Change = packPlanData[i].Plans[1] - packPlanData[i].Plan
			} else {
				packPlanData[i].Change = 0
			}
		}

		// get inventory
		cur, err = s.mgdb.Collection("pack").Find(context.Background(), bson.M{"type": "Inventory"}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(2))
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var packInventoryData []struct {
			Prodtype     string    `bson:"prodtype" json:"prodtype"`
			Inventory    float64   `bson:"inventory" json:"inventory"`
			CreatedAt    time.Time `bson:"createdat" json:"createdat"`
			CreatedAtStr string    `json:"createdatstr"`
		}

		if err := cur.All(context.Background(), &packInventoryData); err != nil {
			log.Println(err)
		}

		for i := 0; i < len(packInventoryData); i++ {
			packInventoryData[i].CreatedAtStr = packInventoryData[i].CreatedAt.Add(7 * time.Hour).Format("15h04 date 2/1")
		}

		// get target
		cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
			// {{"$match", bson.M{"name": "packing total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
			{{"$match", bson.M{"name": "packing total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var packTarget []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err = cur.All(context.Background(), &packTarget); err != nil {
			log.Println(err)
		}
		// get time of latest update
		sr := s.mgdb.Collection("pack").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
		if sr.Err() != nil {
			log.Println(sr.Err())
		}
		var LastReport struct {
			Createdat time.Time `bson:"createdat" json:"createdat"`
		}
		if err := sr.Decode(&LastReport); err != nil {
			log.Println(err)
		}
		packUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

		template.Must(template.ParseFiles("templates/pages/dashboard/pack_valuechart.html")).Execute(w, map[string]interface{}{
			"packData":          packData,
			"packPlanData":      packPlanData,
			"packInventoryData": packInventoryData,
			"packTarget":        packTarget,
			"packUpTime":        packUpTime,
		})

	case "efficiency":
		cur, err := s.mgdb.Collection("pack").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$group", bson.M{"_id": "$date", "value": bson.M{"$sum": "$value"}}}},
			{{"$sort", bson.M{"_id": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var packData []struct {
			Date  string  `bson:"date" json:"date"`
			Value float64 `bson:"value" json:"value"`
		}
		if err := cur.All(context.Background(), &packData); err != nil {
			log.Println(err)
		}
		//get manhr of pack
		cur, err = s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"section": "packing", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.M{"date": 1}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		var packManhr []struct {
			Date   string  `bson:"date" json:"date"`
			HC     int     `bson:"hc" json:"hc"`
			Workhr float64 `bson:"workhr" json:"workhr"`
		}
		if err = cur.All(context.Background(), &packManhr); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/packing_efficiencychart.html")).Execute(w, map[string]interface{}{
			"packData":  packData,
			"packManhr": packManhr,
		})

	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/woodrecovery/getchart
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dwr_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("woodrecoverycharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("woodrecoveryFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("woodrecoveryToDate"))
	log.Println(fromdate, todate, pickedChart)
	switch pickedChart {
	case "general":
		cur, err := s.mgdb.Collection("woodrecovery").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
			{{"$sort", bson.D{{"date", 1}, {"prodtype", 1}}}},
			{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var woodrecoveryData []struct {
			Date     string  `bson:"date" json:"date"`
			Prodtype string  `bson:"prodtype" json:"prodtype"`
			Rate     float64 `bson:"rate" json:"rate"`
		}
		if err := cur.All(context.Background(), &woodrecoveryData); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/dashboard/woodrecovery_generalchart.html")).Execute(w, map[string]interface{}{
			"woodrecoveryData": woodrecoveryData,
		})
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/pack/getchart - change chart of pack area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) ds_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/quality/getchart - change chart of quality area in dashboard
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dq_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickedChart := r.FormValue("qualitycharttype")
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("qualityFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("qualityToDate"))

	switch pickedChart {
	case "general":
		cur, err := s.mgdb.Collection("quality").Aggregate(context.Background(), mongo.Pipeline{
			{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": fromdate.Format("2006-01-02")}}, bson.M{"date": bson.M{"$lte": todate.Format("2006-01-02")}}}}}},
			{{"$group", bson.M{"_id": bson.M{"date": "$date", "section": "$section"}, "checkedqty": bson.M{"$sum": "$checkedqty"}, "failedqty": bson.M{"$sum": "$failedqty"}}}},
			{{"$sort", bson.D{{"_id.date", 1}, {"_id.section", 1}}}},
			{{"$set", bson.M{"date": "$_id.date", "section": "$_id.section"}}},
			{{"$unset", "_id"}},
		})
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var qualityChartData []struct {
			Date       string `bson:"date" json:"date"`
			Section    string `bson:"section" json:"section"`
			CheckedQty int    `bson:"checkedqty" json:"checkedqty"`
			FailedQty  int    `bson:"failedqty" json:"failedqty"`
		}
		if err := cur.All(context.Background(), &qualityChartData); err != nil {
			log.Println(err)
		}

		template.Must(template.ParseFiles("templates/pages/dashboard/quality_generalchart.html")).Execute(w, map[string]interface{}{
			"qualityChartData": qualityChartData,
		})
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// /dashboard/downtime/getchart
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dd_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////
// "/dashboard/safety/getchart"
// ////////////////////////////////////////////////////////////////////////////////
func (s *Server) dst_getchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////
// /request
// //////////////////////////////////////////////////////////
func (s *Server) sendRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/sections/rawwood/entry/entry", s.sre_entry)
func (s *Server) sre_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameTk, err := r.Cookie("username")
	if err != nil {
		log.Println(err)
		w.Write([]byte("Phải đăng nhập"))
		return
	}
	date, _ := time.Parse("2006-01-02", r.FormValue("iqcdate"))
	qty, _ := strconv.ParseFloat(r.FormValue("iqcqty"), 64)

	_, err = s.mgdb.Collection("rawwood").InsertOne(context.Background(), bson.M{
		"type": "import", "date": primitive.NewDateTimeFromTime(date), "qty": qty, "unit": "cbm", "reporter": usernameTk.Value,
		"createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
		w.Write([]byte("Thất bại"))
		return
	}
	w.Write([]byte("Thành công"))
}

// router.POST("/sections/rawwood/entry/selection", s.sre_selection)
func (s *Server) sre_selection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameTk, err := r.Cookie("username")
	if err != nil {
		log.Println(err)
		w.Write([]byte("Phải đăng nhập"))
		return
	}
	date, _ := time.Parse("2006-01-02", r.FormValue("selectiondate"))
	brightqty, _ := strconv.ParseFloat(r.FormValue("brightqty"), 64)
	darkqty, _ := strconv.ParseFloat(r.FormValue("darkqty"), 64)

	if r.FormValue("brightqty") != "" {
		_, err = s.mgdb.Collection("rawwood").InsertOne(context.Background(), bson.M{
			"type": "selection", "date": primitive.NewDateTimeFromTime(date), "qty": brightqty, "unit": "cbm", "reporter": usernameTk.Value,
			"woodtone": "light", "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
			w.Write([]byte("Thất bại"))
			return
		}
	}

	if r.FormValue("darkqty") != "" {
		_, err = s.mgdb.Collection("rawwood").InsertOne(context.Background(), bson.M{
			"type": "selection", "date": primitive.NewDateTimeFromTime(date), "qty": darkqty, "unit": "cbm", "reporter": usernameTk.Value,
			"woodtone": "dark", "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
			w.Write([]byte("Thất bại"))
			return
		}
	}

	if r.FormValue("darkqty") == "" && r.FormValue("darkqty") == "" {
		w.Write([]byte("Thất bại"))
		return
	}

	w.Write([]byte("Thành công"))
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/cutting/overview - get page overview of Cutting
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sc_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/cutting/overview/overview.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/cutting/overview/loadwoodremain - get woodremain area of page overview of Cutting
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sco_loadwoodremain(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tempdate, _ := time.Parse("2006-01-02", "2024-10-01")
	cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "wrnote", "wrremain": bson.M{"$gt": 0}, "date": bson.M{"$gte": primitive.NewDateTimeFromTime(tempdate)}}}},
		{{"$group", bson.M{"_id": "$thickness", "value": bson.M{"$sum": "$wrremain"}}}},
		{{"$sort", bson.M{"value": -1}}},
		{{"$set", bson.M{"name": "$_id"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var woodremainChartData []struct {
		Name  int     `bson:"name" json:"name"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &woodremainChartData); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/cutting/overview/woodremain.html")).Execute(w, map[string]interface{}{
		"woodremainChartData": woodremainChartData,
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/cutting/overview/loadwrnote - load wrnote section of overview of Cutting
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sco_loadwrnote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "wrnote"}}},
		{{"$sort", bson.M{"date": -1}}},
		{{"$limit", 20}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var wrnotes []struct {
		WrnoteCode string  `bson:"wrnotecode"`
		WoodType   string  `bson:"woodtype"`
		Thickness  float64 `bson:"thickness"`
		Date       string  `bson:"date"`
		WrnoteQty  float64 `bson:"wrnoteqty"`
		WrRemain   float64 `bson:"wrremain"`
		ProdType   string  `bson:"prodtype"`
	}
	if err := cur.All(context.Background(), &wrnotes); err != nil {
		log.Println(err)
	}
	numberOfWrnotes := len(wrnotes)

	template.Must(template.ParseFiles("templates/pages/sections/cutting/overview/wrnote.html")).Execute(w, map[string]interface{}{
		"wrnotes":         wrnotes,
		"numberOfWrnotes": numberOfWrnotes,
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/cutting/overview/loadreport - load report section of overview of Cutting
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sco_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "report"}}},
		{{"$sort", bson.M{"date": -1}}},
		{{"$limit", 20}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var reports []struct {
		Wrnote     string  `bson:"wrnote"`
		Woodtype   string  `bson:"woodtype"`
		ProdType   string  `bson:"prodtype"`
		Thickness  float64 `bson:"thickness"`
		Date       string  `bson:"date"`
		Qtycbm     float64 `bson:"qtycbm"`
		Reporter   string  `bson:"reporter"`
		Is25reeded bool    `bson:"is25reeded" json:"is25reeded"`
	}
	if err := cur.All(context.Background(), &reports); err != nil {
		log.Println(err)
	}
	numberOfReports := len(reports)
	totalcbm := 0.0
	for _, v := range reports {
		totalcbm += v.Qtycbm
	}

	template.Must(template.ParseFiles("templates/pages/sections/cutting/overview/report.html")).Execute(w, map[string]interface{}{
		"reports":         reports,
		"numberOfReports": numberOfReports,
		"totalcbm":        fmt.Sprintf("%.3f", totalcbm),
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/cutting/overview/wrnotesearch - search wrnote of overview of Cutting
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sco_wrnotesearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wrnoteseach := r.FormValue("wrnotesearch")
	regexWord := ".*" + wrnoteseach + ".*"
	searchFilter := r.FormValue("searchFilter")

	cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "wrnote", searchFilter: bson.M{"$regex": regexWord, "$options": "i"}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})

	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var wrnotes []struct {
		WrnoteCode string  `bson:"wrnotecode"`
		WoodType   string  `bson:"woodtype"`
		ProdType   string  `bson:"prodtype"`
		Thickness  float64 `bson:"thickness"`
		Date       string  `bson:"date"`
		WrnoteQty  float64 `bson:"wrnoteqty"`
		WrRemain   float64 `bson:"wrremain"`
	}
	if err = cur.All(context.Background(), &wrnotes); err != nil {
		log.Println(err)
	}

	numberOfWrnotes := len(wrnotes)

	template.Must(template.ParseFiles("templates/pages/sections/cutting/overview/wrnote_tbl.html")).Execute(w, map[string]interface{}{
		"wrnotes":         wrnotes,
		"numberOfWrnotes": numberOfWrnotes,
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/cutting/overview/reportsearch - search report of overview of Cutting
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sco_reportsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportsearch := r.FormValue("reportsearch")
	regexWord := ".*" + reportsearch + ".*"
	searchFilter := r.FormValue("searchFilter")

	cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "report", searchFilter: bson.M{"$regex": regexWord, "$options": "i"}}}},
		{{"$sort", bson.M{"date": -1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var reports []struct {
		Wrnote     string  `bson:"wrnote"`
		Woodtype   string  `bson:"woodtype"`
		ProdType   string  `bson:"prodtype"`
		Thickness  float64 `bson:"thickness"`
		Date       string  `bson:"date"`
		Qtycbm     float64 `bson:"qtycbm"`
		Reporter   string  `bson:"reporter"`
		Is25reeded bool    `bson:"is25reeded"`
	}
	if err := cur.All(context.Background(), &reports); err != nil {
		log.Println(err)
	}
	numberOfReports := len(reports)
	totalcbm := 0.0
	for _, v := range reports {
		totalcbm += v.Qtycbm
	}
	template.Must(template.ParseFiles("templates/pages/sections/cutting/overview/report_tbl.html")).Execute(w, map[string]interface{}{
		"reports":         reports,
		"numberOfReports": numberOfReports,
		"totalcbm":        fmt.Sprintf("%.3f", totalcbm),
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/cutting/overview/wrremainfilter - filter remain wrnote of overview of Cutting
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sco_wrnotefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wrnotefilter := r.FormValue("wrnotefilter")

	var filter bson.M
	switch wrnotefilter {
	case "all":
		filter = bson.M{"type": "wrnote"}
	case "undone":
		filter = bson.M{"type": "wrnote", "wrremain": bson.M{"$gt": 0}}
	case "done":
		filter = bson.M{"type": "wrnote", "wrremain": 0}
	}

	cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", filter}},
		{{"$sort", bson.M{"date": -1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})

	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var wrnotes []struct {
		WrnoteCode string  `bson:"wrnotecode"`
		WoodType   string  `bson:"woodtype"`
		ProdType   string  `bson:"prodtype"`
		Thickness  float64 `bson:"thickness"`
		Date       string  `bson:"date"`
		WrnoteQty  float64 `bson:"wrnoteqty"`
		WrRemain   float64 `bson:"wrremain"`
	}
	if err = cur.All(context.Background(), &wrnotes); err != nil {
		log.Println(err)
	}

	numberOfWrnotes := len(wrnotes)

	template.Must(template.ParseFiles("templates/pages/sections/cutting/overview/wrnote_tbl.html")).Execute(w, map[string]interface{}{
		"wrnotes":         wrnotes,
		"numberOfWrnotes": numberOfWrnotes,
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/cutting/overview/reportfilter - filter report of overview of Cutting
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sco_reportfilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportFromDate, _ := time.Parse("2006-01-02", r.FormValue("reportFromDate"))
	reportToDate, _ := time.Parse("2006-01-02", r.FormValue("reportToDate"))

	filter := bson.M{"type": "report", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(reportFromDate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(reportToDate)}}}}

	cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", filter}},
		{{"$sort", bson.M{"date": -1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})

	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var reports []struct {
		Wrnote     string  `bson:"wrnote"`
		Woodtype   string  `bson:"woodtype"`
		ProdType   string  `bson:"prodtype"`
		Thickness  float64 `bson:"thickness"`
		Date       string  `bson:"date"`
		Qtycbm     float64 `bson:"qtycbm"`
		Reporter   string  `bson:"reporter"`
		Is25reeded bool    `bson:"is25reeded"`
	}
	if err = cur.All(context.Background(), &reports); err != nil {
		log.Println(err)
	}

	numberOfReports := len(reports)
	totalcbm := 0.0
	for _, v := range reports {
		totalcbm += v.Qtycbm
	}

	template.Must(template.ParseFiles("templates/pages/sections/cutting/overview/report_tbl.html")).Execute(w, map[string]interface{}{
		"reports":         reports,
		"numberOfReports": numberOfReports,
		"totalcbm":        fmt.Sprintf("%.3f", totalcbm),
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// router.POST("/sections/cutting/overview/createdemand", s.sco_createdemand)
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sco_createdemand(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	startdate, _ := time.Parse("2006-01-02", r.FormValue("cutingdemanstartdate"))
	enddate, _ := time.Parse("2006-01-02", r.FormValue("cutingdemanenddate"))
	reeded25target, _ := strconv.ParseFloat(r.FormValue("cutting25reededdemand"), 64)
	otherstarget, _ := strconv.ParseFloat(r.FormValue("cuttingothersdemand"), 64)

	if r.FormValue("cuttingdemandname") == "" || r.FormValue("cutingdemanstartdate") == "" || r.FormValue("cutingdemanenddate") == "" || r.FormValue("cutting25reededdemand") == "" || r.FormValue("cuttingothersdemand") == "" {
		w.Write([]byte("thiếu thông tin để tạo demand"))
		return
	}

	_, err := s.mgdb.Collection("cutting").UpdateOne(context.Background(), bson.M{"type": "target", "name": r.FormValue("cuttingdemandname")}, bson.M{
		"$set": bson.M{"type": "target", "name": r.FormValue("cuttingdemandname"), "startdate": primitive.NewDateTimeFromTime(startdate), "enddate": primitive.NewDateTimeFromTime(enddate),
			"detail": bson.A{bson.M{"type": "25 Reeded", "target": reeded25target}, bson.M{"type": "Còn lại", "target": otherstarget}}},
	}, options.Update().SetUpsert(true))
	if err != nil {
		log.Println(err)
		return
	}

	// template.Must(template.ParseFiles("templates/pages/sections/cutting/overview/report_tbl.html")).Execute(w, map[string]interface{}{
	// 	"reports":         reports,
	// 	"numberOfReports": numberOfReports,
	// })
}

// //////////////////////////////////////////////////////////
//
//	router.PUT("/sections/cutting/overview/wrnotereturn/:wrnotecode", s.sco_wrnotereturn)
//
// //////////////////////////////////////////////////////////
func (s *Server) sco_wrnotereturn(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameTk, err := r.Cookie("username")
	log.Println(usernameTk.Value)
	if err != nil || usernameTk.Value != "hue" {
		log.Println(err)
		w.Write([]byte("Cần thẩm quyền"))
		return
	}
	wrnotecode := ps.ByName("wrnotecode")
	remainqty, _ := strconv.ParseFloat(ps.ByName("remainqty"), 64)

	result := s.mgdb.Collection("cutting").FindOneAndUpdate(context.Background(), bson.M{"type": "wrnote", "wrnotecode": wrnotecode}, bson.M{"$set": bson.M{"wrremain": 0}})
	if result.Err() != nil {
		log.Println(result.Err())
		w.Write([]byte("Cập nhật thất bại"))
		return
	}
	var cuttingWrnote struct {
		WrnoteCode string    `bson:"wrnotecode"`
		WoodType   string    `bson:"woodtype"`
		Thickness  float64   `bson:"thickness"`
		Date       time.Time `bson:"date"`
		WrnoteQty  float64   `bson:"wrnoteqty"`
		WrRemain   float64   `bson:"wrremain"`
		ProdType   string    `bson:"prodtype"`
		DateStr    string
	}
	if err := result.Decode(&cuttingWrnote); err != nil {
		log.Println(err)
	}
	cuttingWrnote.WrRemain = 0
	cuttingWrnote.DateStr = cuttingWrnote.Date.Format("02-01-2006")

	// create a report return wrnote remain
	_, err = s.mgdb.Collection("cutting").InsertOne(context.Background(), bson.M{
		"type": "wrnote return", "returnwrnote": wrnotecode, "returnqty": remainqty, "reporter": usernameTk.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
		w.Write([]byte("Thất bại"))
		return
	}

	template.Must(template.ParseFiles("templates/pages/sections/cutting/overview/wrnote_tr.html")).Execute(w, map[string]interface{}{
		"cuttingWrnote": cuttingWrnote,
	})
}

// //////////////////////////////////////////////////////////
// /incentive/evaluate/searchstaff
// //////////////////////////////////////////////////////////
func (s *Server) searchstaff(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////
// /incentive/evaluate/searchcriterion
// //////////////////////////////////////////////////////////
func (s *Server) searchcriterion(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////
// /incentive/evaluate/sendevaluate
// //////////////////////////////////////////////////////////
func (s *Server) sendevaluate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////
// "/incentive/admin/
// //////////////////////////////////////////////////////////
func (s *Server) iadmin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/incentive/admin/admin.html",
		"templates/pages/incentive/admin/criteria.html",
		"templates/pages/incentive/admin/evaluate.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////
// "/incentive/admin/loadcrittable
// access collection criteria get criteria data
// then load to criteria table
// //////////////////////////////////////////////////////////
func (s *Server) loadcrittable(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var criteria []models.Criterion
	criteria, err := models.NewCriterionModel(s.mgdb).Find()
	if err != nil {
		log.Println("loadcrittable: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to access criterion collection"))
		return
	}

	template.Must(template.ParseFiles("templates/pages/incentive/admin/crit_table.html")).Execute(w, map[string]interface{}{
		"criteria": criteria,
	})
}

// //////////////////////////////////////////////////////////
// "/incentive/admin/loadevaltable
// access collection evaluation to get evaluate data
// then load to evaluate table
// //////////////////////////////////////////////////////////
func (s *Server) loadevaltable(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////
// /incentive/admin/caupsertcriteria
// upsert a criteria
// then reload criteria table
// //////////////////////////////////////////////////
func (s *Server) caupsertcriteria(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////
// /incentive/admin/deletecriteria
// delete a criteria by id when click on delete icon
// return nothing
// //////////////////////////////////////////////////
func (s *Server) deletecriteria(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////
// /incentive/admin/deleteevaluate
// delete a evaluate by id when click on delete icon
// return nothing
// //////////////////////////////////////////////////
func (s *Server) deleteevaluate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rawevaluateid := ps.ByName("evaluateid")
	evaluateid, _ := primitive.ObjectIDFromHex(rawevaluateid)

	_, err := s.mgdb.Collection("evaluation").DeleteOne(context.Background(), bson.M{"_id": evaluateid})
	if err != nil {
		log.Println("deleteevaluate: ", err)
		w.Write([]byte("Failed to access database"))
		return
	}
}

// //////////////////////////////////////////////////
// /incentive/evaluate
// get page evaluate
// //////////////////////////////////////////////////
func (s *Server) evaluate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/incentive/evaluate/evaluate.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////
// /incentive/admin/searchcriterion - post
// search criteria in admin page
// //////////////////////////////////////////////////
func (s *Server) ia_searchcriterion(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////
// /incentive/admin/searchevaluate - post
// search evaluate in admin page
// //////////////////////////////////////////////////
func (s *Server) ia_searchevaluate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////
// /incentive/overview - get page incentive overview
// //////////////////////////////////////////////////
func (s *Server) ioverview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /incentive/overview/getscorecard - load page incentive overview
// ///////////////////////////////////////////////////////////////////////
func (s *Server) io_loadscores(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /incentive/overview/scoresearch - load point tbody when search
// ///////////////////////////////////////////////////////////////////////
func (s *Server) io_scoresearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /incentive/overview/io_evalsearch - load evaluations tbody when search
// ///////////////////////////////////////////////////////////////////////
func (s *Server) io_evalsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// router.GET("/hr/overview", s.hr_overview)
// ///////////////////////////////////////////////////////////////////////
func (s *Server) hr_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// router.GET("/hr/overview/loadchart", s.hr_loadchart)
// ///////////////////////////////////////////////////////////////////////
func (s *Server) hr_loadchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /hr/admin - load page admin of HR
// ///////////////////////////////////////////////////////////////////////
func (s *Server) hradmin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /hr/admin/searchemployee - search employee in hr admin page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) ha_searchemployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /hr/admin/upsertemployee - update or insert employee into database
// ///////////////////////////////////////////////////////////////////////
func (s *Server) ha_upsertemployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /hr/admin/ha_exportempexcel - create employee list excel file
// ///////////////////////////////////////////////////////////////////////
func (s *Server) ha_exportempexcel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /hr/admin/prevnext - get employee list when click previous, next page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) ha_prevnext(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /hr/entry - get entry page of HR
// ///////////////////////////////////////////////////////////////////////
func (s *Server) hr_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /hr/entry - post to multibly upsert employee list
// ///////////////////////////////////////////////////////////////////////
func (s *Server) hr_insertemplist(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////
// /sections/cutting/entry - get entry page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sc_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("cutting").Find(context.Background(), bson.M{"type": "wrnote", "wrremain": bson.M{"$gt": 0}}, options.Find().SetSort(bson.M{"wrnotecode": 1}))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var wrnotes []struct {
		WrnoteCode string  `bson:"wrnotecode"`
		WrnoteQty  float64 `bson:"wrnoteqty"`
	}
	if err = cur.All(context.Background(), &wrnotes); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/entry.html", "templates/shared/navbar.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": false,
		"wrnotes":           wrnotes,
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/cutting/entry/wrnoteinfo - get wrnote info when select wrnote code
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sc_wrnoteinfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sresult := s.mgdb.Collection("cutting").FindOne(context.Background(), bson.M{"type": "wrnote", "wrnotecode": r.FormValue("wrnote")})
	if sresult.Err() != nil {
		log.Println(sresult.Err())
	}

	var wrnoteinfo struct {
		WrnoteCode string  `bson:"wrnotecode"`
		ProdType   string  `bson:"prodtype"`
		WoodType   string  `bson:"woodtype"`
		WrnoteQty  float64 `bson:"wrnoteqty"`
		Thickness  float64 `bson:"thickness"`
		WrRemain   float64 `bson:"wrremain"`
	}

	if err := sresult.Decode(&wrnoteinfo); err != nil {
		log.Println(err)
	}
	showForReeded := false
	if wrnoteinfo.Thickness == 25 {
		showForReeded = true
	}
	template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/wrnoteinfo.html")).Execute(w, map[string]interface{}{
		"wrnoteinfo":    wrnoteinfo,
		"showForReeded": showForReeded,
	})
}

func (s *Server) sc_newwrnote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/wrnoteinput.html")).Execute(w, nil)
}

func (s *Server) sc_createwrnote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wrnotedate, _ := time.Parse("2006-01-02", r.FormValue("occurdate"))
	prodtype := r.FormValue("prodtype")

	code := r.FormValue("wrnotecode")
	woodtype := r.FormValue("woodtype")
	thickness, _ := strconv.ParseFloat(r.FormValue("thickness"), 64)
	wrnoteqty, _ := strconv.ParseFloat(r.FormValue("wrnoteqty"), 64)
	if code == "" || woodtype == "" || prodtype == "" || thickness == 0 || wrnoteqty == 0 {
		template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/wrnoteinput.html")).Execute(w, map[string]interface{}{
			"showSuccessDialog": false,
			"showMissingDialog": true,
		})
		return
	}
	_, err := s.mgdb.Collection("cutting").InsertOne(context.Background(), bson.M{
		"type": "wrnote", "wrnotecode": code, "prodtype": prodtype, "wrnoteqty": wrnoteqty, "wrremain": wrnoteqty, "woodtype": woodtype, "thickness": thickness, "date": primitive.NewDateTimeFromTime(wrnotedate), "createat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/wrnoteinput.html")).Execute(w, map[string]interface{}{
			"showSuccessDialog": false,
			"showMissingDialog": true,
		})
		return
	}

	cur, err := s.mgdb.Collection("cutting").Find(context.Background(), bson.M{"type": "wrnote", "wrremain": bson.M{"$gt": 0}}, options.Find().SetSort(bson.M{"wrnotecode": 1}))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var wrnotes []struct {
		WrnoteCode string  `bson:"wrnotecode"`
		WrnoteQty  float64 `bson:"wrnoteqty"`
		WrRemain   float64 `bson:"wrremain"`
	}

	if err = cur.All(context.Background(), &wrnotes); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/report_form.html")).Execute(w, map[string]interface{}{
		"wrnotes":           wrnotes,
		"showSuccessDialog": true,
	})
}

// ///////////////////////////////////////////////////////////////////////
// /sections/cutting/sendentry - post entry to database
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sc_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	remain, _ := strconv.ParseFloat(strings.Split(r.FormValue("wrnoteqty"), "/")[0], 64)
	qty, _ := strconv.ParseFloat(r.FormValue("qty"), 64)
	occurdate, _ := time.Parse("2006-01-02", r.FormValue("occurdate"))
	woodtype := r.FormValue("woodtype")
	prodtype := r.FormValue("prodtype")
	thickness, _ := strconv.ParseFloat(r.FormValue("thickness"), 64)
	wrnote := r.FormValue("wrnote")
	forreeded, _ := strconv.ParseBool(r.FormValue("forreeded"))

	usernameToken, err := r.Cookie("username")
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
		return
	}

	if qty == 0 || thickness == 0 || wrnote == "" || qty > remain {
		w.Write([]byte("Sai thông tin nhập liệu"))
		return
	}

	_, err = s.mgdb.Collection("cutting").InsertOne(context.Background(), bson.M{
		"type": "report", "wrnote": wrnote, "woodtype": woodtype, "prodtype": prodtype, "qtycbm": qty, "thickness": thickness, "reporter": usernameToken.Value,
		"date": primitive.NewDateTimeFromTime(occurdate), "createddate": primitive.NewDateTimeFromTime(time.Now()), "lastmodified": primitive.NewDateTimeFromTime(time.Now()),
		"is25reeded": forreeded,
	})
	if err != nil {
		log.Println(err)
	}

	// update remain qty of wrnote
	_, err = s.mgdb.Collection("cutting").UpdateOne(context.Background(), bson.M{"type": "wrnote", "wrnotecode": wrnote}, bson.M{"$set": bson.M{"wrremain": math.Round((remain-qty)*1000) / 1000}})
	if err != nil {
		log.Println(err)
	}

	http.Redirect(w, r, "/sections/cutting/entry", http.StatusSeeOther)
}

// ///////////////////////////////////////////////////////////////////////
//
//	router.GET("/sections/cutting/entry/return", s.sce_return)
//
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sce_return(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/return.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// router.POST("/sections/cutting/entry/sendreturn", s.sce_sendreturn)
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sce_sendreturn(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	returndate, _ := time.Parse("2006-01-02", r.FormValue("returndate"))
	wrnotecode := r.FormValue("wrnotecode")
	returnqty, _ := strconv.ParseFloat(r.FormValue("returnqty"), 64)
	returntype := false
	if r.FormValue("returntype") == "true" {
		returntype = true
	}

	usernameToken, err := r.Cookie("username")
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
		return
	}

	if r.FormValue("returnqty") == "" || r.FormValue("returntype") == "" {
		w.Write([]byte("Thiếu thông tin nhập liệu"))
		return
	}

	_, err = s.mgdb.Collection("cutting").InsertOne(context.Background(), bson.M{
		"type": "return", "wrnote": wrnotecode, "qtycbm": returnqty, "reporter": usernameToken.Value,
		"date": primitive.NewDateTimeFromTime(returndate), "createdat": primitive.NewDateTimeFromTime(time.Now()),
		"is25": returntype,
	})
	if err != nil {
		log.Println(err)
	}

	http.Redirect(w, r, "/sections/cutting/entry/return", http.StatusSeeOther)
}

// ///////////////////////////////////////////////////////////////////////
//
//	router.GET("/sections/cutting/entry/fine", s.sce_fine)
//
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sce_fine(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/fine.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// router.POST("/sections/cutting/entry/sendfine", s.sce_sendfine)
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sce_sendfine(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	date, _ := time.Parse("2006-01-02", r.FormValue("finedate"))
	qty, _ := strconv.ParseFloat(r.FormValue("fineqty"), 64)
	is25reeded := false
	if r.FormValue("finetype") == "true" {
		is25reeded = true
	}
	usernameToken, err := r.Cookie("username")
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
		return
	}

	if r.FormValue("fineqty") == "" || r.FormValue("finetype") == "" {
		w.Write([]byte("Thiếu thông tin nhập liệu"))
		return
	}

	_, err = s.mgdb.Collection("cutting").InsertOne(context.Background(), bson.M{
		"type": "fine", "qtycbm": qty, "reporter": usernameToken.Value, "is25reeded": is25reeded,
		"date": primitive.NewDateTimeFromTime(date), "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
	}

	http.Redirect(w, r, "/sections/cutting/entry/fine", http.StatusSeeOther)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/cutting/admin - get page admin of cutting section
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sc_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/cutting/admin/admin.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/cutting/admin/loadreports - load report area on cutting admin page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sc_loadreports(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("cutting").Find(context.Background(), bson.M{"type": "report"}, options.Find().SetSort(bson.M{"createddate": -1}).SetLimit(20))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var cuttingReports []struct {
		ReportId     string    `bson:"_id"`
		Date         time.Time `bson:"date"`
		Wrnote       string    `bson:"wrnote"`
		Woodtype     string    `bson:"woodtype"`
		ProdType     string    `bson:"prodtype"`
		Thickness    float64   `bson:"thickness"`
		Qty          float64   `bson:"qtycbm"`
		Type         string    `bson:"type"`
		Reporter     string    `bson:"reporter"`
		CreatedDate  time.Time `bson:"createddate"`
		LastModified time.Time `bson:"lastmodified"`
	}
	if err := cur.All(context.Background(), &cuttingReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/cutting/admin/reports.html")).Execute(w, map[string]interface{}{
		"cuttingReports":  cuttingReports,
		"numberOfReports": len(cuttingReports),
	})
}

// ///////////////////////////////////////////////////////////////////////
// /sections/cutting/admin/loadwrnote - load wrnote area on cutting admin page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sc_loadwrnote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("cutting").Find(context.Background(), bson.M{"type": "wrnote"}, options.Find().SetSort(bson.M{"createat": -1}).SetLimit(20))
	if err != nil {
		log.Println(err)
		return
	}
	defer cur.Close(context.Background())
	var cuttingWrnote []struct {
		WrnoteId    string    `bson:"_id"`
		WrnoteCode  string    `bson:"wrnotecode"`
		Woodtype    string    `bson:"woodtype"`
		ProdType    string    `bson:"prodtype"`
		Thickness   float64   `bson:"thickness"`
		Qty         float64   `bson:"wrnoteqty"`
		Remain      float64   `bson:"wrremain"`
		Date        time.Time `bson:"date"`
		CreatedDate time.Time `bson:"createat"`
	}
	if err := cur.All(context.Background(), &cuttingWrnote); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/cutting/admin/wrnotes.html")).Execute(w, map[string]interface{}{
		"cuttingWrnotes":  cuttingWrnote,
		"numberOfWrnotes": len(cuttingWrnote),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/cutting/admin/deletereport/:reportid - delete a report on page admin of cutting section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sca_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))

	deletedReport := s.mgdb.Collection("cutting").FindOneAndDelete(context.Background(), bson.M{"_id": reportid})
	if deletedReport.Err() != nil {
		log.Println(deletedReport.Err())
		return
	}
	var report struct {
		ReportId     string    `bson:"_id"`
		Date         time.Time `bson:"date"`
		Wrnote       string    `bson:"wrnote"`
		Woodtype     string    `bson:"woodtype"`
		Thickness    float64   `bson:"thickness"`
		Qty          float64   `bson:"qtycbm"`
		Type         string    `bson:"type"`
		Reporter     string    `bson:"reporter"`
		CreatedDate  time.Time `bson:"createddate"`
		LastModified time.Time `bson:"lastmodified"`
	}
	if err := deletedReport.Decode(&report); err != nil {
		log.Println(err)
		return
	}
	// return quantity for wrnote
	wrnote := s.mgdb.Collection("cutting").FindOneAndUpdate(context.Background(), bson.M{"type": "wrnote", "wrnotecode": report.Wrnote},
		bson.M{"$inc": bson.M{"wrremain": math.Round(report.Qty*1000) / 1000}})
	if wrnote.Err() != nil {
		log.Println(wrnote.Err())
		return
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/cutting/admin/deletewrnote/:wrnoteid - delete a wrnote on page admin of cutting section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sca_deletewrnote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wrnoteid, _ := primitive.ObjectIDFromHex(ps.ByName("wrnoteid"))

	_, err := s.mgdb.Collection("cutting").DeleteOne(context.Background(), bson.M{"_id": wrnoteid})
	if err != nil {
		log.Println(err)
		return
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/cutting/admin/wrnoteupdateform/:wrnoteid - update a wrnote on page admin of cutting section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sca_wrnoteupdateform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wrnotecode := ps.ByName("wrnoteid")

	result := s.mgdb.Collection("cutting").FindOne(context.Background(), bson.M{"type": "wrnote", "wrnotecode": wrnotecode})
	if result.Err() != nil {
		log.Println(result.Err())
		return
	}
	var cuttingWrnote struct {
		WrnoteId    string    `bson:"_id"`
		WrnoteCode  string    `bson:"wrnotecode"`
		Woodtype    string    `bson:"woodtype"`
		ProdType    string    `bson:"prodtype"`
		Thickness   float64   `bson:"thickness"`
		Qty         float64   `bson:"wrnoteqty"`
		Remain      float64   `bson:"wrremain"`
		Date        time.Time `bson:"date"`
		CreatedDate time.Time `bson:"createat"`
	}
	if err := result.Decode(&cuttingWrnote); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/cutting/admin/wrnoteupdate_form.html")).Execute(w, map[string]interface{}{
		"cuttingWrnote": cuttingWrnote,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/cutting/admin/updatewrnote/:wrnoteid - update a wrnote on page admin of cutting section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sca_updatewrnote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wrnotecode := ps.ByName("wrnoteid")

	prodtype := r.FormValue("prodtype")
	date, _ := time.Parse("2006-01-02", r.FormValue("occurdate"))
	qtycbm, _ := strconv.ParseFloat(r.FormValue("qtycbm"), 64)

	cur, err := s.mgdb.Collection("cutting").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"type": "report", "wrnote": wrnotecode}}},
		{{"$group", bson.M{"_id": "$wrnote", "qty": bson.M{"$sum": "$qtycbm"}}}},
		{{"$set", bson.M{"wrnote": "$_id"}}},
	})
	if err != nil {
		log.Println(err)
	}
	var alreadyCut []struct {
		Wrnotecode string  `bson:"wrnote"`
		Qty        float64 `bson:"qty"`
	}
	if err := cur.All(context.Background(), &alreadyCut); err != nil {
		log.Println(err)
	}
	if qtycbm < alreadyCut[0].Qty {
		w.Write([]byte("Thất bại. Nấu số cập nhập mới nhỏ hơn số cbm đã cắt thì phải xóa báo cáo cắt trước"))
		return
	}
	qtyremain := (math.Round(qtycbm*1000) - math.Round(alreadyCut[0].Qty*1000)) / 1000

	result := s.mgdb.Collection("cutting").FindOneAndUpdate(context.Background(), bson.M{"wrnotecode": wrnotecode}, bson.M{"$set": bson.M{
		"wrnoteqty": qtycbm, "wrremain": qtyremain, "prodtype": prodtype, "date": primitive.NewDateTimeFromTime(date)}})
	if result.Err() != nil {
		log.Println(result.Err())
		return
	}
	var cuttingWrnote struct {
		WrnoteId    string    `bson:"_id"`
		WrnoteCode  string    `bson:"wrnotecode"`
		Woodtype    string    `bson:"woodtype"`
		ProdType    string    `bson:"prodtype"`
		Thickness   float64   `bson:"thickness"`
		Qty         float64   `bson:"wrnoteqty"`
		Remain      float64   `bson:"wrremain"`
		Date        time.Time `bson:"date"`
		CreatedDate time.Time `bson:"createat"`
	}
	if err := result.Decode(&cuttingWrnote); err != nil {
		log.Println(err)
	}
	cuttingWrnote.ProdType = prodtype
	cuttingWrnote.Date = date
	cuttingWrnote.Qty = qtycbm
	cuttingWrnote.Remain = qtyremain

	// update reports
	_, err = s.mgdb.Collection("cutting").UpdateMany(context.Background(), bson.M{"type": "report", "wrnote": cuttingWrnote.WrnoteCode}, bson.M{"$set": bson.M{"prodtype": prodtype}})
	if err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/cutting/admin/wrnote_tr.html")).Execute(w, map[string]interface{}{
		"cuttingWrnote": cuttingWrnote,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/cutting/admin/searchreport - search reports on page admin of cutting section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sca_searchreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cuttingReports := models.NewCuttingModel(s.mgdb).Search(r.FormValue("reportSearch"))
	template.Must(template.ParseFiles("templates/pages/sections/cutting/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"cuttingReports": cuttingReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/cutting/admin/earchwrnote - search wrnote on page admin of cutting section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sca_searchwrnote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cuttingWrnotes := models.NewCuttingModel(s.mgdb).WrnoteSearch(r.FormValue("wrnoteSearch"))
	template.Must(template.ParseFiles("templates/pages/sections/cutting/admin/wrnote_tbody.html")).Execute(w, map[string]interface{}{
		"cuttingWrnotes": cuttingWrnotes,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/cutting/admin/reportdatefilter
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sca_reportdatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("cuttingFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("cuttingToDate"))

	cur, err := s.mgdb.Collection("cutting").Find(context.Background(), bson.M{
		"type": "report",
		"$and": bson.A{
			bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}},
			bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}},
		},
	}, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println("failed to find cutting report at sca_reportdatefilter")
	}
	defer cur.Close(context.Background())
	var cuttingReports []struct {
		ReportId     string    `bson:"_id"`
		Date         time.Time `bson:"date"`
		Wrnote       string    `bson:"wrnote"`
		ProdType     string    `bson:"prodtype"`
		Woodtype     string    `bson:"woodtype"`
		Thickness    float64   `bson:"thickness"`
		Qty          float64   `bson:"qtycbm"`
		Type         string    `bson:"type"`
		Reporter     string    `bson:"reporter"`
		CreatedDate  time.Time `bson:"createddate"`
		LastModified time.Time `bson:"lastmodified"`
	}
	if err := cur.All(context.Background(), &cuttingReports); err != nil {
		log.Println("failed to decode cuttingReports at sca_reportdatefilter")
	}
	template.Must(template.ParseFiles("templates/pages/sections/cutting/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"cuttingReports": cuttingReports,
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/lamination/overview - get page overview of Lamination
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sl_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/lamination/overview/overview.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/lamination/overview/loadreport - load report table of page overview of Lamination
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) slo_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("lamination").Aggregate(context.Background(), mongo.Pipeline{
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "at": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var laminationReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		ProdType    string  `bson:"prodtype"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"at"`
	}
	if err := cur.All(context.Background(), &laminationReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/lamination/overview/report.html")).Execute(w, map[string]interface{}{
		"laminationReports": laminationReports,
		"numberOfReports":   len(laminationReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/lamination/admin/searchreport - search reports on page admin of lamination section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) slo_reportsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportsearch := r.FormValue("reportsearch")
	regexWord := ".*" + reportsearch + ".*"
	searchFilter := r.FormValue("searchFilter")

	cur, err := s.mgdb.Collection("lamination").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{searchFilter: bson.M{"$regex": regexWord, "$options": "i"}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m/%Y", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var laminationReports []struct {
		Date      string  `bson:"date"`
		ProdType  string  `bson:"prodtype"`
		Qty       float64 `bson:"qty"`
		Reporter  string  `bson:"reporter"`
		CreatedAt string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &laminationReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/lamination/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"laminationReports": laminationReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/lamination/overview/reportdatefilter
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) slo_reportdatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("laminationFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("laminationToDate"))

	cur, err := s.mgdb.Collection("lamination").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m/%Y", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var laminationReports []struct {
		Date      string  `bson:"date"`
		ProdType  string  `bson:"prodtype"`
		Qty       float64 `bson:"qty"`
		Reporter  string  `bson:"reporter"`
		CreatedAt string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &laminationReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/lamination/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"laminationReports": laminationReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/lamination/entry/entry - load page entry of lamination section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sl_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/lamination/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/lamination/entry/loadform - load form of page entry of lamination section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sle_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/lamination/entry/form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/lamination/entry/sendentry - post form of page entry of lamination section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sle_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	qty, _ := strconv.ParseFloat(r.FormValue("qty"), 64)
	prodtype := r.FormValue("prodtype")
	log.Println(qty)
	if r.FormValue("prodtype") == "" || r.FormValue("qty") == "" {
		template.Must(template.ParseFiles("templates/pages/sections/lamination/entry/form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	_, err := s.mgdb.Collection("lamination").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "prodtype": prodtype, "qty": qty, "createdat": primitive.NewDateTimeFromTime(time.Now()), "reporter": username,
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/lamination/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/sections/lamination/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// ///////////////////////////////////////////////////////////////////////
// /sections/lamination/admin - get page admin of lamination section
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sl_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/lamination/admin/admin.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/lamination/admin/loadreport - load report area on lamination admin page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sla_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("lamination").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(50))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var laminationReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		ProdType    string    `bson:"prodtype"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &laminationReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/lamination/admin/report.html")).Execute(w, map[string]interface{}{
		"laminationReports": laminationReports,
		"numberOfReports":   len(laminationReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/lamination/admin/searchreport - search reports on page admin of lamination section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sla_searchreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchWord := r.FormValue("reportSearch")
	regexWord := ".*" + searchWord + ".*"
	dateSearch, err := time.Parse("2006-01-02", searchWord)
	var filter bson.M
	if err != nil {
		filter = bson.M{"$or": bson.A{
			bson.M{"reporter": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"prodtype": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"qty": bson.M{"$regex": regexWord, "$options": "i"}},
		},
		}
	} else {
		filter = bson.M{"date": primitive.NewDateTimeFromTime(dateSearch)}
	}
	cur, err := s.mgdb.Collection("lamination").Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var laminationReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		ProdType    string    `bson:"prodtype"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &laminationReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/lamination/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"laminationReports": laminationReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/lamination/admin/deletereport/:reportid - delete a report on page admin of lamination section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sla_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))

	_, err := s.mgdb.Collection("lamination").DeleteOne(context.Background(), bson.M{"_id": reportid})
	if err != nil {
		log.Println(err)
		return
	}
}

// router.GET("/sections/slicing/entry", s.ss_entry)
func (s *Server) ss_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/slicing/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// router.GET("/sections/slicing/entry/loadform", s.sse_loadform)
func (s *Server) sse_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/slicing/entry/form.html")).Execute(w, nil)
}

// outer.POST("/sections/slicing/entry/sendentry", s.sse_sendentry)
func (s *Server) sse_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	qty, _ := strconv.ParseFloat(r.FormValue("qty"), 64)
	prodtype := r.FormValue("prodtype")
	log.Println(qty)
	if r.FormValue("prodtype") == "" || r.FormValue("qty") == "" {
		template.Must(template.ParseFiles("templates/pages/sections/slicing/entry/form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	_, err := s.mgdb.Collection("slicing").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "prodtype": prodtype, "qty": qty, "createdat": primitive.NewDateTimeFromTime(time.Now()), "reporter": username,
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/slicing/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/sections/slicing/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// router.GET("/sections/slicing/admin", s.ss_admin)
func (s *Server) ss_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/slicing/admin/admin.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// router.GET("/sections/slicing/admin/loadreport", s.ssa_loadreport)
func (s *Server) ssa_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("slicing").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -3))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{
			"date":      bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}},
			"createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M %d-%m-%Y", "date": "$createdat"}},
		}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var slicingData []struct {
		Id        string  `bson:"_id"`
		Date      string  `bson:"date"`
		CreatedAt string  `bson:"createdat"`
		ProdType  string  `bson:"prodtype"`
		Qty       float64 `bson:"qty"`
		Reporter  string  `bson:"reporter"`
	}
	if err := cur.All(context.Background(), &slicingData); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/slicing/admin/report.html")).Execute(w, map[string]interface{}{
		"slicingData": slicingData,
	})
}

// router.POST("/sections/slicing/admin/reportsearch", s.ssa_reportsearch)
func (s *Server) ssa_reportsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchRegex := ".*" + r.FormValue("reportsearch") + ".*"

	cur, err := s.mgdb.Collection("slicing").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$or": bson.A{
			bson.M{"prodtype": bson.M{"$regex": searchRegex, "$options": "i"}},
			bson.M{"report": bson.M{"$regex": searchRegex, "$options": "i"}},
		}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{
			"date":      bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}},
			"createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M %d-%m-%Y", "date": "$createdat"}},
		}}},
	})

	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var slicingData []struct {
		Id        string  `bson:"_id"`
		Date      string  `bson:"date"`
		CreatedAt string  `bson:"createdat"`
		ProdType  string  `bson:"prodtype"`
		Qty       float64 `bson:"qty"`
		Reporter  string  `bson:"reporter"`
	}
	if err := cur.All(context.Background(), &slicingData); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/slicing/admin/table.html")).Execute(w, map[string]interface{}{
		"slicingData": slicingData,
	})
}

// router.DELETE("/sections/slicing/admin/deletereport/:id", s.ssa_deletereport)
func (s *Server) ssa_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("id"))

	_, err := s.mgdb.Collection("slicing").DeleteOne(context.Background(), bson.M{"_id": reportid})
	if err != nil {
		log.Println(err)
		return
	}
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/reededline/overview - get page overview of reededline
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sr_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/reededline/overview/overview.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/reededline/overview/loadreport - load report table of page overview of reededline
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sro_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("reededline").Aggregate(context.Background(), mongo.Pipeline{
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "at": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var reededlineReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Tone        string  `bson:"tone"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"at"`
	}
	if err := cur.All(context.Background(), &reededlineReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/reededline/overview/report.html")).Execute(w, map[string]interface{}{
		"reededlineReports": reededlineReports,
		"numberOfReports":   len(reededlineReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/reededline/admin/searchreport - search reports on page admin of reededline section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sro_reportsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportsearch := r.FormValue("reportsearch")
	regexWord := ".*" + reportsearch + ".*"
	searchFilter := r.FormValue("searchFilter")

	cur, err := s.mgdb.Collection("reededline").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{searchFilter: bson.M{"$regex": regexWord, "$options": "i"}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "at": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var reededlineReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Tone        string  `bson:"tone"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"at"`
	}
	if err = cur.All(context.Background(), &reededlineReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/reededline/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"reededlineReports": reededlineReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/reededline/overview/reportdatefilter
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sro_reportdatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("reededlineFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("reededlineToDate"))

	cur, err := s.mgdb.Collection("reededline").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "at": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var reededlineReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Tone        string  `bson:"tone"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"at"`
	}
	if err = cur.All(context.Background(), &reededlineReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/reededline/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"reededlineReports": reededlineReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/reedeline/entry - load page entry of reededline section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sr_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/reededline/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/reedeline/entry/loadform - load form of page entry of reededline section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sre_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/reededline/entry/form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/reedeline/entry/sendentry - load form of page entry of reededline section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sre_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	qty, _ := strconv.ParseFloat(r.FormValue("qty"), 64)
	prodtype := r.FormValue("tone")
	if r.FormValue("tone") == "" || r.FormValue("qty") == "" {
		template.Must(template.ParseFiles("templates/pages/sections/reededline/entry/form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	_, err := s.mgdb.Collection("reededline").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "tone": prodtype, "qty": qty, "createdat": primitive.NewDateTimeFromTime(time.Now()), "reporter": username,
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/reededline/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/sections/reededline/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// ///////////////////////////////////////////////////////////////////////
// /sections/reededline/admin - get page admin of reededline section
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sr_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/reededline/admin/admin.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/reededline/admin/loadreport - load report area on reededline admin page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sra_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("reededline").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(50))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var reededlineReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Tone        string    `bson:"tone"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &reededlineReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/reededline/admin/report.html")).Execute(w, map[string]interface{}{
		"reededlineReports": reededlineReports,
		"numberOfReports":   len(reededlineReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/reededline/admin/searchreport - search reports on page admin of reededline section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sra_searchreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchWord := r.FormValue("reportSearch")
	regexWord := ".*" + searchWord + ".*"
	dateSearch, err := time.Parse("2006-01-02", searchWord)
	var filter bson.M
	if err != nil {
		filter = bson.M{"$or": bson.A{
			bson.M{"tone": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"reporter": bson.M{"$regex": regexWord, "$options": "i"}},
		},
		}
	} else {
		filter = bson.M{"date": primitive.NewDateTimeFromTime(dateSearch)}
	}
	cur, err := s.mgdb.Collection("reededline").Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var reededlineReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Tone        string    `bson:"tone"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &reededlineReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/reededline/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"reededlineReports": reededlineReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/reededline/admin/deletereport/:reportid - delete a report on page admin of reededline section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sra_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))

	_, err := s.mgdb.Collection("reededline").DeleteOne(context.Background(), bson.M{"_id": reportid})
	if err != nil {
		log.Println(err)
		return
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/output/entry - load page entry of output section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) so_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/output/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/output/entry/loadentry
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) soe_loadentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/output/entry/form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/output/entry/sendentry
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) soe_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	qty, _ := strconv.ParseFloat(r.FormValue("qty"), 64)
	outputtype := r.FormValue("outputtype")
	section := r.FormValue("section")
	if r.FormValue("outputtype") == "" || r.FormValue("qty") == "" || r.FormValue("section") == "" {
		template.Must(template.ParseFiles("templates/pages/sections/output/entry/form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	_, err := s.mgdb.Collection("output").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "type": outputtype, "section": section, "qty": qty, "createdat": primitive.NewDateTimeFromTime(time.Now()), "reporter": username,
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/output/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/sections/output/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/output/entry/loadformentry
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) soe_loadformentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/output/entry/fastform.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/output/entry/sendfastentry
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) soe_sendfastentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	outputtype := r.FormValue("outputtype")
	outputlistraw := strings.Fields(r.FormValue("outputlist"))
	var outputlist = make([]float64, len(outputlistraw))
	for i := 0; i < len(outputlistraw); i++ {
		var a float64
		a, err := strconv.ParseFloat(outputlistraw[i], 64)
		if err != nil {
			log.Println(err)
			template.Must(template.ParseFiles("templates/pages/sections/output/entry/form.html")).Execute(w, map[string]interface{}{
				"showErrDialog": true,
				"msgDialog":     "Phải nhập chuỗi số.",
			})
			return
		}
		outputlist[i] = a
	}
	if r.FormValue("outputtype") == "" || r.FormValue("outputlist") == "" {
		template.Must(template.ParseFiles("templates/pages/sections/output/entry/fastform.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	var bdoc []interface{}
	var firsection = []string{"1.Slice", "2.Selection", "3.Lamination", "9.Delivery"}
	var reededsection = []string{"1.Slice", "2.Selection", "3.Lamination", "4.Drying", "5.Reeding", "6.Selection-2", "7.Tubi", "8.Veneer"}
	if outputtype == "fir" && len(outputlist) == 4 {
		for i := 0; i < len(firsection); i++ {
			b := bson.M{
				"date": primitive.NewDateTimeFromTime(date), "type": outputtype, "section": firsection[i], "qty": outputlist[i], "createdat": primitive.NewDateTimeFromTime(time.Now()), "reporter": username,
			}
			bdoc = append(bdoc, b)
		}
	}
	if outputtype == "reeded" && len(outputlist) == 8 {
		for i := 0; i < len(reededsection); i++ {
			b := bson.M{
				"date": primitive.NewDateTimeFromTime(date), "type": outputtype, "section": reededsection[i], "qty": outputlist[i], "createdat": primitive.NewDateTimeFromTime(time.Now()), "reporter": username,
			}
			bdoc = append(bdoc, b)
		}
	}
	_, err := s.mgdb.Collection("output").InsertMany(context.Background(), bdoc)
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/output/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/sections/output/entry/fastform.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// ///////////////////////////////////////////////////////////////////////
// /sections/output/admin
// ///////////////////////////////////////////////////////////////////////
func (s *Server) so_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/output/admin/admin.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/output/admin/loadreport
// ///////////////////////////////////////////////////////////////////////
func (s *Server) soa_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("output").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(20))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var outputReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Type        string    `bson:"type"`
		Section     string    `bson:"section"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &outputReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/output/admin/report.html")).Execute(w, map[string]interface{}{
		"outputReports":   outputReports,
		"numberOfReports": len(outputReports),
	})
}

// ///////////////////////////////////////////////////////////////////////
// /sections/output/admin/loadreport
// ///////////////////////////////////////////////////////////////////////
func (s *Server) soa_searchreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchRegex := ".*" + r.FormValue("reportSearch") + ".*"
	// searchNumber, _ := strconv.ParseFloat(r.FormValue("reportSearch"), 64)

	cur, err := s.mgdb.Collection("output").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$or": bson.A{
			bson.M{"section": bson.M{"$regex": searchRegex, "$options": "i"}},
			bson.M{"type": bson.M{"$regex": searchRegex, "$options": "i"}},
			// bson.M{"qty": searchNumber},
		}}}},
		{{"$sort", bson.M{"date": -1}}},
	})
	if err != nil {
		log.Println("failed to access output at soa_searchreport")
	}
	defer cur.Close(context.Background())
	var outputReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Type        string    `bson:"type"`
		Section     string    `bson:"section"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}

	if err := cur.All(context.Background(), &outputReports); err != nil {
		log.Println("failed to decode at soa_searchreport")
	}
	template.Must(template.ParseFiles("templates/pages/sections/output/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"outputReports": outputReports,
	})
}

// ///////////////////////////////////////////////////////////////////////
// /sections/output/admin/reportdatefilter
// ///////////////////////////////////////////////////////////////////////
func (s *Server) soa_reportdatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("fromdate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("todate"))

	cur, err := s.mgdb.Collection("output").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{
			bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}},
			bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}},
		}}}},
		{{"$sort", bson.M{"date": -1}}},
	})
	if err != nil {
		log.Println("failed to access output at soa_searchreport")
	}
	defer cur.Close(context.Background())
	var outputReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Type        string    `bson:"type"`
		Section     string    `bson:"section"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}

	if err := cur.All(context.Background(), &outputReports); err != nil {
		log.Println("failed to decode at soa_searchreport")
	}
	template.Must(template.ParseFiles("templates/pages/sections/output/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"outputReports": outputReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/output/admin/deletereport/:reportid
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) soa_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))

	_, err := s.mgdb.Collection("output").DeleteOne(context.Background(), bson.M{"_id": reportid})
	if err != nil {
		log.Println(err)
		return
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/output/admin/updateform/:reportid
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) soa_updateform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))

	result := s.mgdb.Collection("output").FindOne(context.Background(), bson.M{"_id": id})
	if result.Err() != nil {
		log.Println(result.Err())
		return
	}
	var outputReports struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Type        string    `bson:"type"`
		Section     string    `bson:"section"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := result.Decode(&outputReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/output/admin/update_form.html")).Execute(w, map[string]interface{}{
		"outputReports": outputReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// "/sections/output/admin/updatereport/:reportid"
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) soa_updatereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))
	outputtype := r.FormValue("outputtype")
	section := r.FormValue("section")
	qty, _ := strconv.ParseFloat(r.FormValue("qty"), 64)

	result := s.mgdb.Collection("output").FindOneAndUpdate(context.Background(), bson.M{"_id": id}, bson.M{"$set": bson.M{"type": outputtype, "section": section, "qty": qty}})
	if result.Err() != nil {
		log.Println(result.Err())
		return
	}
	var outputReports struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Type        string    `bson:"type"`
		Section     string    `bson:"section"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := result.Decode(&outputReports); err != nil {
		log.Println(err)
	}
	outputReports.Qty = qty
	outputReports.Type = outputtype
	outputReports.Section = section

	template.Must(template.ParseFiles("templates/pages/sections/output/admin/updated_tr.html")).Execute(w, map[string]interface{}{
		"outputReports": outputReports,
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/veneer/overview - get page overview of veneer
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sv_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/veneer/overview/overview.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/veneer/overview/loadreport - load report table of page overview of veneer
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) svo_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("veneer").Aggregate(context.Background(), mongo.Pipeline{
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var veneerReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Type        string  `bson:"type"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &veneerReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/veneer/overview/report.html")).Execute(w, map[string]interface{}{
		"veneerReports":   veneerReports,
		"numberOfReports": len(veneerReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/veneer/admin/searchreport - search reports on page admin of veneer section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) svo_reportsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportsearch := r.FormValue("reportsearch")
	regexWord := ".*" + reportsearch + ".*"
	searchFilter := r.FormValue("searchFilter")

	cur, err := s.mgdb.Collection("veneer").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{searchFilter: bson.M{"$regex": regexWord, "$options": "i"}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var veneerReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Type        string  `bson:"type"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &veneerReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/veneer/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"veneerReports": veneerReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/veneer/overview/reportdatefilter
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) svo_reportdatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("veneerFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("veneerToDate"))

	cur, err := s.mgdb.Collection("veneer").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var veneerReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Type        string  `bson:"type"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &veneerReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/veneer/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"veneerReports": veneerReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/veneer/entry - load page entry of veneer section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sv_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/veneer/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/veneer/entry/loadform - load form of page entry of veneer section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sve_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/veneer/entry/form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/veneer/admin/searchreport - search reports on page admin of veneer section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sva_searchreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchWord := r.FormValue("reportSearch")
	regexWord := ".*" + searchWord + ".*"
	dateSearch, err := time.Parse("2006-01-02", searchWord)
	var filter bson.M
	if err != nil {
		filter = bson.M{"$or": bson.A{
			bson.M{"type": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"reporter": bson.M{"$regex": regexWord, "$options": "i"}},
		},
		}
	} else {
		filter = bson.M{"date": primitive.NewDateTimeFromTime(dateSearch)}
	}
	cur, err := s.mgdb.Collection("veneer").Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var veneerReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Type        string    `bson:"type"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &veneerReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/veneer/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"veneerReports": veneerReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/veneer/entry/sendentry - load form of page entry of veneer section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sve_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	qty, _ := strconv.ParseFloat(r.FormValue("qty"), 64)
	veneertype := r.FormValue("type")
	if r.FormValue("type") == "" || r.FormValue("qty") == "" {
		template.Must(template.ParseFiles("templates/pages/sections/veneer/entry/form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	_, err := s.mgdb.Collection("veneer").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "type": veneertype, "qty": qty, "createdat": primitive.NewDateTimeFromTime(time.Now()), "reporter": username,
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/veneer/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/sections/veneer/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// ///////////////////////////////////////////////////////////////////////
// /sections/veneer/admin - get page admin of veneer section
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sv_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/veneer/admin/admin.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/veneer/admin/loadreport - load report area on veneer admin page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sva_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("veneer").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(50))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var veneerReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Type        string    `bson:"type"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &veneerReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/veneer/admin/report.html")).Execute(w, map[string]interface{}{
		"veneerReports":   veneerReports,
		"numberOfReports": len(veneerReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/veneer/admin/deletereport/:reportid - delete a report on page admin of veneer section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sva_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))

	_, err := s.mgdb.Collection("veneer").DeleteOne(context.Background(), bson.M{"_id": reportid})
	if err != nil {
		log.Println(err)
		return
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/veneer/entry - load page entry of veneer section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sf_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/finemill/entry/loadform - load form of page entry of finemill section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sfe_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/finemill/entry/form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/finemill/sendentry - post form of page entry of finemill section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sfe_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/assembly/overview - get page overview of assembly
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sa_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/assembly/overview/overview.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/assembly/overview/loadreport - load report table of page overview of assembly
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sao_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var assemblyReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Value       float64 `bson:"value"`
		ProdType    string  `bson:"prodtype"`
		Itemcode    string  `bson:"itemcode"`
		ItemType    string  `bson:"itemtype"`
		Component   string  `bson:"component"`
		Factory     string  `bson:"factory"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &assemblyReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/assembly/overview/report.html")).Execute(w, map[string]interface{}{
		"assemblyReports": assemblyReports,
		"numberOfReports": len(assemblyReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/assembly/admin/searchreport - search reports on page admin of assembly section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sao_reportsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportsearch := r.FormValue("reportsearch")
	regexWord := ".*" + reportsearch + ".*"
	searchFilter := r.FormValue("searchFilter")

	cur, err := s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{searchFilter: bson.M{"$regex": regexWord, "$options": "i"}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var assemblyReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Value       float64 `bson:"value"`
		ProdType    string  `bson:"prodtype"`
		Itemcode    string  `bson:"itemcode"`
		ItemType    string  `bson:"itemtype"`
		Component   string  `bson:"component"`
		Factory     string  `bson:"factory"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &assemblyReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/assembly/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"assemblyReports": assemblyReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/assembly/overview/reportdatefilter
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sao_reportdatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("assemblyFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("assemblyToDate"))

	cur, err := s.mgdb.Collection("assembly").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdad", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var assemblyReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Value       float64 `bson:"value"`
		ProdType    string  `bson:"prodtype"`
		Itemcode    string  `bson:"itemcode"`
		ItemType    string  `bson:"itemtype"`
		Component   string  `bson:"component"`
		Factory     string  `bson:"factory"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &assemblyReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/assembly/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"assemblyReports": assemblyReports,
	})
}

// router.POST("/sections/assembly/overview/addplanvalue", s.sao_addplanvalue)
func (s *Server) sao_addplanvalue(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, err := r.Cookie("username")
	if err != nil {
		w.Write([]byte("Không có thẩm quyền"))
		return
	}
	date, _ := time.Parse("2006-01-02", r.FormValue("plandate"))
	brandplanvalue, _ := strconv.ParseFloat(r.FormValue("brandplanvalue"), 64)
	rhplanvalue, _ := strconv.ParseFloat(r.FormValue("rhplanvalue"), 64)

	// _, err = s.mgdb.Collection("assembly").UpdateOne(context.Background(), bson.D{{"type", "plan"}, {"date", primitive.NewDateTimeFromTime(date)}, {"plantype", "brand"}}, bson.M{
	// 	"$set": bson.M{"type": "plan", "date": primitive.NewDateTimeFromTime(date), "plantype": "brand", "plan": brandplanvalue, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now())},
	// }, options.Update().SetUpsert(true))
	_, err = s.mgdb.Collection("assembly").InsertOne(context.Background(), bson.M{
		"type": "plan", "date": primitive.NewDateTimeFromTime(date), "plantype": "brand", "plan": brandplanvalue, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
	}
	// _, err = s.mgdb.Collection("assembly").UpdateOne(context.Background(), bson.D{{"type", "plan"}, {"date", primitive.NewDateTimeFromTime(date)}, {"plantype", "rh"}}, bson.M{
	// 	"$set": bson.M{"type": "plan", "date": primitive.NewDateTimeFromTime(date), "plantype": "rh", "plan": rhplanvalue, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now())},
	// }, options.Update().SetUpsert(true))
	_, err = s.mgdb.Collection("assembly").InsertOne(context.Background(), bson.M{
		"type": "plan", "date": primitive.NewDateTimeFromTime(date), "plantype": "rh", "plan": rhplanvalue, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// router.POST("/sections/assembly/overview/updateinventory", s.sao_updateinventory)
func (s *Server) sao_updateinventory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, err := r.Cookie("username")
	if err != nil {
		w.Write([]byte("Không có thẩm quyền"))
		return
	}

	x3brandinventory, _ := strconv.ParseFloat(r.FormValue("x3brandinventory"), 64)
	x3rhinventory, _ := strconv.ParseFloat(r.FormValue("x3rhinventory"), 64)
	x7brandinventory, _ := strconv.ParseFloat(r.FormValue("x7brandinventory"), 64)
	x7rhinventory, _ := strconv.ParseFloat(r.FormValue("x7rhinventory"), 64)

	if r.FormValue("x3brandinventory") != "" {
		_, err = s.mgdb.Collection("assembly").InsertOne(context.Background(), bson.M{
			"type": "Inventory", "prodtype": "brand", "factory": "2", "inventory": x3brandinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	if r.FormValue("x3rhinventory") != "" {
		_, err = s.mgdb.Collection("assembly").InsertOne(context.Background(), bson.M{
			"type": "Inventory", "prodtype": "rh", "factory": "2", "inventory": x3rhinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	if r.FormValue("x7brandinventory") != "" {
		_, err = s.mgdb.Collection("assembly").InsertOne(context.Background(), bson.M{
			"type": "Inventory", "prodtype": "brand", "factory": "1", "inventory": x7brandinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	if r.FormValue("x7rhinventory") != "" {
		_, err = s.mgdb.Collection("assembly").InsertOne(context.Background(), bson.M{
			"type": "Inventory", "prodtype": "rh", "factory": "1", "inventory": x7rhinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	s.d_loadassembly(w, r, ps)

}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/assembly/entry - load page entry of assembly section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sa_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/assembly/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/assembly/loadform - load form of page entry of assembly section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sae_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/assembly/entry/form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/assembly/sendentry - post form of page entry of assembly section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sae_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	itemtype := r.FormValue("itemtype")
	// itemtype := "whole"
	// if r.FormValue("switch") != "" {
	// 	itemtype = r.FormValue("switch")
	// }
	itemcode := r.FormValue("itemcode")
	component := r.FormValue("component")
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	factory := r.FormValue("factory")
	prodtype := r.FormValue("prodtype")
	qty, _ := strconv.Atoi(r.FormValue("qty"))
	value, _ := strconv.ParseFloat(r.FormValue("value"), 64)

	if factory == "" || prodtype == "" {
		template.Must(template.ParseFiles("templates/pages/sections/assembly/entry/form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	insertedResult, err := s.mgdb.Collection("assembly").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "itemcode": itemcode, "itemtype": itemtype, "component": component,
		"factory": factory, "prodtype": prodtype, "qty": qty, "value": value, "reporter": username, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/assembly/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}

	//wait to 26th Nov to open
	//create a report for production value collection when White Product were inserted
	if prodtype == "white" {
		_, err = s.mgdb.Collection("prodvalue").InsertOne(context.Background(), bson.M{
			"date": primitive.NewDateTimeFromTime(date), "item": itemcode, "itemtype": itemtype,
			"factory": factory, "prodtype": prodtype, "qty": qty, "value": value, "reporter": username, "createdat": primitive.NewDateTimeFromTime(time.Now()),
			"from": "assembly", "refid": insertedResult.InsertedID,
		})
		if err != nil {
			log.Println(err)
			template.Must(template.ParseFiles("templates/pages/sections/pack/entry/form.html")).Execute(w, map[string]interface{}{
				"showErrDialog": true,
				"msgDialog":     "Không cập nhật được vào prodvalue",
			})
			return
		}
	}

	template.Must(template.ParseFiles("templates/pages/sections/assembly/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// router.GET("/sections/assembly/planentry", s.sae_planentry)
func (s *Server) sae_planentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/assembly/entry/planentry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// router.GET("/sections/assembly/entry/loadplanform", s.sae_loadplanform)
func (s *Server) sae_loadplanform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/assembly/entry/planform.html")).Execute(w, nil)
}

// router.POST("/sections/assembly/entry/sendplanentry", s.sae_sendplanentry)
func (s *Server) sae_sendplanentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value

	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	value, _ := strconv.ParseFloat(r.FormValue("value"), 64)

	if r.FormValue("value") == "" {
		template.Must(template.ParseFiles("templates/pages/sections/assembly/entry/planform.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	_, err := s.mgdb.Collection("assembly").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "plan": value, "reporter": username, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/assembly/entry/planform.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/sections/assembly/entry/planform.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// router.GET("/sections/assembly/inventoryentry", s.sai_inventoryentry)
func (s *Server) sai_inventoryentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/assembly/entry/inventoryentry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// router.GET("/sections/assembly/entry/loadinventoryform", s.sai_loadinventoryform)
func (s *Server) sai_loadinventoryform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/assembly/entry/inventoryform.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/assembly/admin - get page admin of assembly section
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sa_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/assembly/admin/admin.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/assembly/admin/loadreport - load report area on assembly admin page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) saa_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("assembly").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(100))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var assemblyReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Value       float64   `bson:"value"`
		ProdType    string    `bson:"prodtype"`
		Itemcode    string    `bson:"itemcode"`
		ItemType    string    `bson:"itemtype"`
		Component   string    `bson:"component"`
		Factory     string    `bson:"factory"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &assemblyReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/assembly/admin/report.html")).Execute(w, map[string]interface{}{
		"assemblyReports": assemblyReports,
		"numberOfReports": len(assemblyReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/assembly/admin/searchreport - search reports on page admin of assembly section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) saa_searchreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchWord := r.FormValue("reportSearch")
	regexWord := ".*" + searchWord + ".*"
	dateSearch, err := time.Parse("2006-01-02", searchWord)
	var filter bson.M
	if err != nil {
		filter = bson.M{"$or": bson.A{
			bson.M{"itemcode": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"component": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"prodtype": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"itemtype": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"factory": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"reporter": bson.M{"$regex": regexWord, "$options": "i"}},
		},
		}
	} else {
		filter = bson.M{"date": primitive.NewDateTimeFromTime(dateSearch)}
	}
	cur, err := s.mgdb.Collection("assembly").Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var assemblyReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Value       float64   `bson:"value"`
		ProdType    string    `bson:"prodtype"`
		Itemcode    string    `bson:"itemcode"`
		ItemType    string    `bson:"itemtype"`
		Component   string    `bson:"component"`
		Factory     string    `bson:"factory"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &assemblyReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/assembly/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"assemblyReports": assemblyReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/veneer/admin/deletereport/:reportid - delete a report on page admin of veneer section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) saa_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))

	deletedPackReport := s.mgdb.Collection("assembly").FindOneAndDelete(context.Background(), bson.M{"_id": reportid})
	if deletedPackReport.Err() != nil {
		log.Println(deletedPackReport.Err())
		return
	}
	var assemblyReport struct {
		ReportID string `bson:"_id"`
		Prodtype string `bson:"prodtype"`
	}
	if err := deletedPackReport.Decode(&assemblyReport); err != nil {
		log.Println(err)
	}

	if assemblyReport.Prodtype == "white" {
		refidObject, _ := primitive.ObjectIDFromHex(assemblyReport.ReportID)
		// update production value
		result := s.mgdb.Collection("prodvalue").FindOneAndDelete(context.Background(), bson.M{"refid": refidObject})
		if result.Err() != nil {
			log.Println(result.Err())
		}
	}
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/woodfinish/overview - get page overview of assembly
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) sw_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/woodfinish/overview/overview.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/woodfinish/overview/loadreport - load report table of page overview of woodfinish
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) swo_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var woodfinishReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Value       float64 `bson:"value"`
		ProdType    string  `bson:"prodtype"`
		Itemcode    string  `bson:"itemcode"`
		ItemType    string  `bson:"itemtype"`
		Component   string  `bson:"component"`
		Factory     string  `bson:"factory"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &woodfinishReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/woodfinish/overview/report.html")).Execute(w, map[string]interface{}{
		"woodfinishReports": woodfinishReports,
		"numberOfReports":   len(woodfinishReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/woodfinish/admin/searchreport - search reports on page admin of woodfinish section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) swo_reportsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportsearch := r.FormValue("reportsearch")
	regexWord := ".*" + reportsearch + ".*"
	searchFilter := r.FormValue("searchFilter")

	cur, err := s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{searchFilter: bson.M{"$regex": regexWord, "$options": "i"}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var woodfinishReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Value       float64 `bson:"value"`
		ProdType    string  `bson:"prodtype"`
		Itemcode    string  `bson:"itemcode"`
		ItemType    string  `bson:"itemtype"`
		Component   string  `bson:"component"`
		Factory     string  `bson:"factory"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &woodfinishReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/woodfinish/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"woodfinishReports": woodfinishReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/woodfinish/overview/reportdatefilter
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) swo_reportdatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("woodfinishFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("woodfinishToDate"))

	cur, err := s.mgdb.Collection("woodfinish").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var woodfinishReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Value       float64 `bson:"value"`
		ProdType    string  `bson:"prodtype"`
		Itemcode    string  `bson:"itemcode"`
		ItemType    string  `bson:"itemtype"`
		Component   string  `bson:"component"`
		Factory     string  `bson:"factory"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &woodfinishReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/woodfinish/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"woodfinishReports": woodfinishReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/woodfinish/entry - load page entry of woodfinish section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sw_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/woodfinish/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/woodfinish/loadform - load form of page entry of woodfinish section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) swe_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/woodfinish/entry/form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/woodfinish/sendentry - post form of page entry of woodfinish section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) swe_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	itemtype := r.FormValue("itemtype")
	// itemtype := "whole"
	// if r.FormValue("switch") != "" {
	// 	itemtype = r.FormValue("switch")
	// }
	itemcode := r.FormValue("itemcode")
	component := r.FormValue("component")
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	factory := r.FormValue("factory")
	prodtype := r.FormValue("prodtype")
	qty, _ := strconv.Atoi(r.FormValue("qty"))
	value, _ := strconv.ParseFloat(r.FormValue("value"), 64)

	if factory == "" || prodtype == "" {
		template.Must(template.ParseFiles("templates/pages/sections/woodfinish/entry/form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	_, err := s.mgdb.Collection("woodfinish").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "itemcode": itemcode, "itemtype": itemtype, "component": component,
		"factory": factory, "prodtype": prodtype, "qty": qty, "value": value, "reporter": username, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/woodfinish/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/sections/woodfinish/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// router.POST("/sections/woodfinish/overview/addplanvalue", s.swo_addplanvalue)
func (s *Server) swo_addplanvalue(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, err := r.Cookie("username")
	if err != nil {
		w.Write([]byte("Không có thẩm quyền"))
		return
	}
	woodfinishdate, _ := time.Parse("2006-01-02", r.FormValue("woodfinishplandate"))
	woodfinishbrandplanvalue, _ := strconv.ParseFloat(r.FormValue("woodfinishbrandplanvalue"), 64)
	woodfinishrhplanvalue, _ := strconv.ParseFloat(r.FormValue("woodfinishrhplanvalue"), 64)

	_, err = s.mgdb.Collection("woodfinish").InsertOne(context.Background(), bson.M{"type": "plan", "date": primitive.NewDateTimeFromTime(woodfinishdate), "plantype": "brand", "plan": woodfinishbrandplanvalue, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now())})
	if err != nil {
		log.Println(err)
	}
	_, err = s.mgdb.Collection("woodfinish").InsertOne(context.Background(), bson.M{"type": "plan", "date": primitive.NewDateTimeFromTime(woodfinishdate), "plantype": "rh", "plan": woodfinishrhplanvalue, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now())})
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// router.POST("/sections/woodfinish/overview/updateinventory", s.swo_updateinventory)
func (s *Server) swo_updateinventory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, err := r.Cookie("username")
	if err != nil {
		w.Write([]byte("Không có thẩm quyền"))
		return
	}

	x3brandinventory, _ := strconv.ParseFloat(r.FormValue("x3brandinventory"), 64)
	x3rhinventory, _ := strconv.ParseFloat(r.FormValue("x3rhinventory"), 64)
	x7brandinventory, _ := strconv.ParseFloat(r.FormValue("x7brandinventory"), 64)
	x7rhinventory, _ := strconv.ParseFloat(r.FormValue("x7rhinventory"), 64)

	if r.FormValue("x3brandinventory") != "" {
		_, err = s.mgdb.Collection("woodfinish").InsertOne(context.Background(), bson.M{
			"type": "Inventory", "prodtype": "brand", "factory": "2", "inventory": x3brandinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	if r.FormValue("x3rhinventory") != "" {
		_, err = s.mgdb.Collection("woodfinish").InsertOne(context.Background(), bson.M{
			"type": "Inventory", "prodtype": "rh", "factory": "2", "inventory": x3rhinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	if r.FormValue("x7brandinventory") != "" {
		_, err = s.mgdb.Collection("woodfinish").InsertOne(context.Background(), bson.M{
			"type": "Inventory", "prodtype": "brand", "factory": "1", "inventory": x7brandinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	if r.FormValue("x7rhinventory") != "" {
		_, err = s.mgdb.Collection("woodfinish").InsertOne(context.Background(), bson.M{
			"type": "Inventory", "prodtype": "rh", "factory": "1", "inventory": x7rhinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	s.d_loadwoodfinish(w, r, ps)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/woodfinish/admin - get page admin of woodfinish section
// ///////////////////////////////////////////////////////////////////////
func (s *Server) sw_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/woodfinish/admin/admin.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/woodfinish/admin/loadreport - load report area on woodfinish admin page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) swa_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("woodfinish").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(100))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var woodfinishReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Value       float64   `bson:"value"`
		ProdType    string    `bson:"prodtype"`
		Itemcode    string    `bson:"itemcode"`
		ItemType    string    `bson:"itemtype"`
		Component   string    `bson:"component"`
		Factory     string    `bson:"factory"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &woodfinishReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/woodfinish/admin/report.html")).Execute(w, map[string]interface{}{
		"woodfinishReports": woodfinishReports,
		"numberOfReports":   len(woodfinishReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/woodfinish/admin/searchreport - search reports on page admin of woodfinish section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) swa_searchreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchWord := r.FormValue("reportSearch")
	regexWord := ".*" + searchWord + ".*"
	dateSearch, err := time.Parse("2006-01-02", searchWord)
	var filter bson.M
	if err != nil {
		filter = bson.M{"$or": bson.A{
			bson.M{"itemcode": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"component": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"prodtype": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"itemtype": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"factory": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"reporter": bson.M{"$regex": regexWord, "$options": "i"}},
		},
		}
	} else {
		filter = bson.M{"date": primitive.NewDateTimeFromTime(dateSearch)}
	}
	cur, err := s.mgdb.Collection("woodfinish").Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var woodfinishReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Value       float64   `bson:"value"`
		ProdType    string    `bson:"prodtype"`
		Itemcode    string    `bson:"itemcode"`
		ItemType    string    `bson:"itemtype"`
		Component   string    `bson:"component"`
		Factory     string    `bson:"factory"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &woodfinishReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/woodfinish/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"woodfinishReports": woodfinishReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/woodfinish/admin/deletereport/:reportid - delete a report on page admin of woodfinish section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) swa_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))

	_, err := s.mgdb.Collection("woodfinish").DeleteOne(context.Background(), bson.M{"_id": reportid})
	if err != nil {
		log.Println(err)
		return
	}
}

// router.POST("/sections/whitewood/overview/addmoney", s.swo_addmoney)
func (s *Server) swo_addmoney(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameTk, _ := r.Cookie("username")
	date, _ := time.Parse("2006-01-02", r.FormValue("whitewoodmoneydate"))
	brandmoney, _ := strconv.ParseFloat(r.FormValue("whitewoodbrandmoney"), 64)
	rhmoney, _ := strconv.ParseFloat(r.FormValue("whitewoodrhmoney"), 64)
	whitemoney, _ := strconv.ParseFloat(r.FormValue("whitewoodwhitemoney"), 64)

	if r.FormValue("whitewoodbrandmoney") != "" {
		_, err := s.mgdb.Collection("whitewood").InsertOne(context.Background(), bson.M{
			"date": primitive.NewDateTimeFromTime(date), "prodtype": "brand", "value": brandmoney, "reporter": usernameTk.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	if r.FormValue("whitewoodrhmoney") != "" {
		_, err := s.mgdb.Collection("whitewood").InsertOne(context.Background(), bson.M{
			"date": primitive.NewDateTimeFromTime(date), "prodtype": "rh", "value": rhmoney, "reporter": usernameTk.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	if r.FormValue("whitewoodwhitemoney") != "" {
		_, err := s.mgdb.Collection("whitewood").InsertOne(context.Background(), bson.M{
			"date": primitive.NewDateTimeFromTime(date), "prodtype": "white", "value": whitemoney, "reporter": usernameTk.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		})
		if err != nil {
			log.Println(err)
		}
	}

	cur, err := s.mgdb.Collection("whitewood").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.prodtype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": "$_id.prodtype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var whitewoodData []struct {
		Date  string  `bson:"date" json:"date"`
		Type  string  `bson:"type" json:"type"`
		Value float64 `bson:"value" json:"value"`
	}
	if err := cur.All(context.Background(), &whitewoodData); err != nil {
		log.Println(err)
	}

	// get plan data
	cur, err = s.mgdb.Collection("whitewood").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}}}}},
		{{"$sort", bson.M{"createdat": -1}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "plantype": "$plantype"}, "plan": bson.M{"$first": "$plan"}, "plans": bson.M{"$firstN": bson.M{"input": "$plan", "n": 2}}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.plantype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "plantype": "$_id.plantype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var whitewoodPlanData []struct {
		Date     string    `bson:"date" json:"date"`
		Plantype string    `bson:"plantype" json:"plantype"`
		Plan     float64   `bson:"plan" json:"plan"`
		Plans    []float64 `bson:"plans" json:"plans"`
		Change   float64   `json:"change"`
	}

	if err := cur.All(context.Background(), &whitewoodPlanData); err != nil {
		log.Println(err)
	}
	for i := 0; i < len(whitewoodPlanData); i++ {
		if len(whitewoodPlanData[i].Plans) >= 2 && whitewoodPlanData[i].Plans[1] != 0 {
			whitewoodPlanData[i].Change = whitewoodPlanData[i].Plans[1] - whitewoodPlanData[i].Plan
		} else {
			whitewoodPlanData[i].Change = 0
		}
	}

	// get inventory
	cur, err = s.mgdb.Collection("whitewood").Find(context.Background(), bson.M{"type": "Inventory"}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(2))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var whitewoodInventoryData []struct {
		Prodtype     string    `bson:"prodtype" json:"prodtype"`
		Inventory    float64   `bson:"inventory" json:"inventory"`
		CreatedAt    time.Time `bson:"createdat" json:"createdat"`
		CreatedAtStr string    `json:"createdatstr"`
	}

	if err := cur.All(context.Background(), &whitewoodInventoryData); err != nil {
		log.Println(err)
	}

	for i := 0; i < len(whitewoodInventoryData); i++ {
		whitewoodInventoryData[i].CreatedAtStr = whitewoodInventoryData[i].CreatedAt.Add(7 * time.Hour).Format("15h04 date 2/1")
	}
	// get target
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"name": "whitewood total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -15))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var whitewoodTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &whitewoodTarget); err != nil {
		log.Println(err)
	}

	// get time of latest update
	sr := s.mgdb.Collection("whitewood").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var LastReport struct {
		Createdat time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&LastReport); err != nil {
		log.Println(err)
	}
	whitewoodUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

	template.Must(template.ParseFiles("templates/pages/dashboard/whitewood_generalchart.html")).Execute(w, map[string]interface{}{
		"whitewoodData":          whitewoodData,
		"whitewoodPlanData":      whitewoodPlanData,
		"whitewoodInventoryData": whitewoodInventoryData,
		"whitewoodTarget":        whitewoodTarget,
		"whitewoodUpTime":        whitewoodUpTime,
	})
}

// router.POST("/sections/whitewood/overview/addplanvalue", s.swwo_addplanvalue)
func (s *Server) swwo_addplanvalue(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, err := r.Cookie("username")
	if err != nil {
		w.Write([]byte("Không có thẩm quyền"))
		return
	}
	date, _ := time.Parse("2006-01-02", r.FormValue("whitewoodplandate"))
	brandplanvalue, _ := strconv.ParseFloat(r.FormValue("whitewoodbrandplanvalue"), 64)
	rhplanvalue, _ := strconv.ParseFloat(r.FormValue("whitewoodrhplanvalue"), 64)

	_, err = s.mgdb.Collection("whitewood").InsertOne(context.Background(), bson.M{"type": "plan", "date": primitive.NewDateTimeFromTime(date), "plantype": "brand", "plan": brandplanvalue, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now())})
	if err != nil {
		log.Println(err)
	}
	_, err = s.mgdb.Collection("whitewood").InsertOne(context.Background(), bson.M{"type": "plan", "date": primitive.NewDateTimeFromTime(date), "plantype": "rh", "plan": rhplanvalue, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now())})
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// router.POST("/sections/whitewood/overview/updateinventory", s.swwo_updateinventory)
func (s *Server) swwo_updateinventory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, err := r.Cookie("username")
	if err != nil {
		w.Write([]byte("Không có thẩm quyền"))
		return
	}

	brandinventory, _ := strconv.ParseFloat(r.FormValue("whitewoodbrandinventory"), 64)
	rhinventory, _ := strconv.ParseFloat(r.FormValue("whitewoodrhinventory"), 64)

	_, err = s.mgdb.Collection("whitewood").InsertOne(context.Background(), bson.M{
		"type": "Inventory", "prodtype": "rh", "inventory": brandinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
	}

	_, err = s.mgdb.Collection("whitewood").InsertOne(context.Background(), bson.M{
		"type": "Inventory", "prodtype": "brand", "inventory": rhinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// router.POST("/sections/whitewood/overview/addnammoney", s.swo_addnammoney)
func (s *Server) swo_addnammoney(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameTk, _ := r.Cookie("username")
	date, _ := time.Parse("2006-01-02", r.FormValue("whitewoodnamdate"))
	value, _ := strconv.ParseFloat(r.FormValue("whitewoodnammoney"), 64)

	if r.FormValue("whitewoodnammoney") == "" {
		w.Write([]byte("Thiếu thông tin"))
		return
	}

	_, err := s.mgdb.Collection("whitewood").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "prodtype": "variance", "value": value, "reporter": usernameTk.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
	}

	// load chart
	cur, err := s.mgdb.Collection("whitewood").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": bson.M{"$exists": false}}, bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.prodtype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "type": "$_id.prodtype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var whitewoodData []struct {
		Date  string  `bson:"date" json:"date"`
		Type  string  `bson:"type" json:"type"`
		Value float64 `bson:"value" json:"value"`
	}
	if err := cur.All(context.Background(), &whitewoodData); err != nil {
		log.Println(err)
	}

	// get plan data
	cur, err = s.mgdb.Collection("whitewood").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"type": "plan", "date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -10))}}}}}},
		{{"$sort", bson.M{"createdat": -1}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "plantype": "$plantype"}, "plan": bson.M{"$first": "$plan"}, "plans": bson.M{"$firstN": bson.M{"input": "$plan", "n": 2}}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.plantype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id.date"}}, "plantype": "$_id.plantype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var whitewoodPlanData []struct {
		Date     string    `bson:"date" json:"date"`
		Plantype string    `bson:"plantype" json:"plantype"`
		Plan     float64   `bson:"plan" json:"plan"`
		Plans    []float64 `bson:"plans" json:"plans"`
		Change   float64   `json:"change"`
	}

	if err := cur.All(context.Background(), &whitewoodPlanData); err != nil {
		log.Println(err)
	}
	for i := 0; i < len(whitewoodPlanData); i++ {
		if len(whitewoodPlanData[i].Plans) >= 2 && whitewoodPlanData[i].Plans[1] != 0 {
			whitewoodPlanData[i].Change = whitewoodPlanData[i].Plans[1] - whitewoodPlanData[i].Plan
		} else {
			whitewoodPlanData[i].Change = 0
		}
	}

	// get inventory
	cur, err = s.mgdb.Collection("whitewood").Find(context.Background(), bson.M{"type": "Inventory"}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(2))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var whitewoodInventoryData []struct {
		Prodtype     string    `bson:"prodtype" json:"prodtype"`
		Inventory    float64   `bson:"inventory" json:"inventory"`
		CreatedAt    time.Time `bson:"createdat" json:"createdat"`
		CreatedAtStr string    `json:"createdatstr"`
	}

	if err := cur.All(context.Background(), &whitewoodInventoryData); err != nil {
		log.Println(err)
	}

	for i := 0; i < len(whitewoodInventoryData); i++ {
		whitewoodInventoryData[i].CreatedAtStr = whitewoodInventoryData[i].CreatedAt.Add(7 * time.Hour).Format("15h04 date 2/1")
	}
	// get target
	cur, err = s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"name": "whitewood total by date", "$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -15))}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now())}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var whitewoodTarget []struct {
		Date  string  `bson:"date" json:"date"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &whitewoodTarget); err != nil {
		log.Println(err)
	}

	// get time of latest update
	sr := s.mgdb.Collection("whitewood").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"createdat": -1}))
	if sr.Err() != nil {
		log.Println(sr.Err())
	}
	var LastReport struct {
		Createdat time.Time `bson:"createdat" json:"createdat"`
	}
	if err := sr.Decode(&LastReport); err != nil {
		log.Println(err)
	}
	whitewoodUpTime := LastReport.Createdat.Add(7 * time.Hour).Format("15:04")

	template.Must(template.ParseFiles("templates/pages/dashboard/whitewood_generalchart.html")).Execute(w, map[string]interface{}{
		"whitewoodData":          whitewoodData,
		"whitewoodPlanData":      whitewoodPlanData,
		"whitewoodInventoryData": whitewoodInventoryData,
		"whitewoodTarget":        whitewoodTarget,
		"whitewoodUpTime":        whitewoodUpTime,
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/pack/overview - get page overview of assembly
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) spk_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/pack/overview/overview.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/pack/overview/loadreport - load report table of page overview of pack
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) pko_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("pack").Aggregate(context.Background(), mongo.Pipeline{
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var packReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Value       float64 `bson:"value"`
		ProdType    string  `bson:"prodtype"`
		Itemcode    string  `bson:"itemcode"`
		ItemType    string  `bson:"itemtype"`
		Part        string  `bson:"part"`
		Factory     string  `bson:"factory"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &packReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/pack/overview/report.html")).Execute(w, map[string]interface{}{
		"packReports":     packReports,
		"numberOfReports": len(packReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/pack/admin/searchreport - search reports on page admin of pack section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) pko_reportsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportsearch := r.FormValue("reportsearch")
	regexWord := ".*" + reportsearch + ".*"
	searchFilter := r.FormValue("searchFilter")

	cur, err := s.mgdb.Collection("pack").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{searchFilter: bson.M{"$regex": regexWord, "$options": "i"}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var packReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Value       float64 `bson:"value"`
		ProdType    string  `bson:"prodtype"`
		Itemcode    string  `bson:"itemcode"`
		ItemType    string  `bson:"itemtype"`
		Part        string  `bson:"part"`
		Factory     string  `bson:"factory"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &packReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/pack/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"packReports": packReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/pack/overview/reportdatefilter
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) pko_reportdatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("packingFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("packingToDate"))
	cur, err := s.mgdb.Collection("pack").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var packReports []struct {
		ReportId    string  `bson:"_id"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		Value       float64 `bson:"value"`
		ProdType    string  `bson:"prodtype"`
		Itemcode    string  `bson:"itemcode"`
		ItemType    string  `bson:"itemtype"`
		Part        string  `bson:"part"`
		Factory     string  `bson:"factory"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &packReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/pack/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"packReports": packReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/pack/entry - load page entry of pack section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spk_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/pack/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/pack/loadform - load form of page entry of pack section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spk_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/pack/entry/form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/pack/sendentry - post form of page entry of pack section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spk_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	itemtype := r.FormValue("itemtype")
	// itemtype := "whole"
	// if r.FormValue("switch") != "" {
	// 	itemtype = r.FormValue("switch")
	// }
	itemcode := r.FormValue("itemcode")
	part := r.FormValue("part")
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	factory := r.FormValue("factory")
	prodtype := r.FormValue("prodtype")
	qty, _ := strconv.Atoi(r.FormValue("qty"))
	value, _ := strconv.ParseFloat(r.FormValue("value"), 64)

	if factory == "" || prodtype == "" {
		template.Must(template.ParseFiles("templates/pages/sections/pack/entry/form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	insertedResult, err := s.mgdb.Collection("pack").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "itemcode": itemcode, "itemtype": itemtype, "part": part,
		"factory": factory, "prodtype": prodtype, "qty": qty, "value": value, "reporter": username, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/pack/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}

	//create a report for production value collection
	if prodtype != "stock" {
		_, err = s.mgdb.Collection("prodvalue").InsertOne(context.Background(), bson.M{
			"date": primitive.NewDateTimeFromTime(date), "item": itemcode, "itemtype": itemtype,
			"factory": factory, "prodtype": prodtype, "qty": qty, "value": value, "reporter": username, "createdat": primitive.NewDateTimeFromTime(time.Now()),
			"from": "pack", "refid": insertedResult.InsertedID,
		})
		if err != nil {
			log.Println(err)
			template.Must(template.ParseFiles("templates/pages/sections/pack/entry/form.html")).Execute(w, map[string]interface{}{
				"showErrDialog": true,
				"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
			})
			return
		}
	}

	template.Must(template.ParseFiles("templates/pages/sections/pack/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// router.POST("/sections/pack/overview/updateinventory", s.spo_updateinventory)
func (s *Server) spo_updateinventory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, err := r.Cookie("username")
	if err != nil {
		w.Write([]byte("Không có thẩm quyền"))
		return
	}

	packbrandinventory, _ := strconv.ParseFloat(r.FormValue("packbrandinventory"), 64)
	packrhinventory, _ := strconv.ParseFloat(r.FormValue("packrhinventory"), 64)

	_, err = s.mgdb.Collection("pack").InsertOne(context.Background(), bson.M{
		"type": "Inventory", "prodtype": "rh", "inventory": packrhinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
	}

	_, err = s.mgdb.Collection("pack").InsertOne(context.Background(), bson.M{
		"type": "Inventory", "prodtype": "brand", "inventory": packbrandinventory, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// router.POST("/sections/pack/overview/addplanvalue", s.spo_addplanvalue)
func (s *Server) spo_addplanvalue(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, err := r.Cookie("username")
	if err != nil {
		w.Write([]byte("Không có thẩm quyền"))
		return
	}
	date, _ := time.Parse("2006-01-02", r.FormValue("packplandate"))
	packbrandplanvalue, _ := strconv.ParseFloat(r.FormValue("packbrandplanvalue"), 64)
	packrhplanvalue, _ := strconv.ParseFloat(r.FormValue("packrhplanvalue"), 64)

	_, err = s.mgdb.Collection("pack").InsertOne(context.Background(), bson.M{"type": "plan", "date": primitive.NewDateTimeFromTime(date), "plantype": "brand", "plan": packbrandplanvalue, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now())})
	if err != nil {
		log.Println(err)
	}
	_, err = s.mgdb.Collection("pack").InsertOne(context.Background(), bson.M{"type": "plan", "date": primitive.NewDateTimeFromTime(date), "plantype": "rh", "plan": packrhplanvalue, "reporter": usernameToken.Value, "createdat": primitive.NewDateTimeFromTime(time.Now())})
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/pack/admin - get page admin of pack section
// ///////////////////////////////////////////////////////////////////////
func (s *Server) spk_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/pack/admin/admin.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/pack/admin/loadreport - load report area on pack admin page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) spka_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("pack").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(100))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var packReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Value       float64   `bson:"value"`
		ProdType    string    `bson:"prodtype"`
		Itemcode    string    `bson:"itemcode"`
		ItemType    string    `bson:"itemtype"`
		Part        string    `bson:"part"`
		Factory     string    `bson:"factory"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &packReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/pack/admin/report.html")).Execute(w, map[string]interface{}{
		"packReports":     packReports,
		"numberOfReports": len(packReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/pack/admin/searchreport - search reports on page admin of pack section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spka_searchreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchWord := r.FormValue("reportSearch")
	regexWord := ".*" + searchWord + ".*"
	dateSearch, err := time.Parse("2006-01-02", searchWord)
	var filter bson.M
	if err != nil {
		filter = bson.M{"$or": bson.A{
			bson.M{"itemcode": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"part": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"prodtype": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"itemtype": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"factory": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"reporter": bson.M{"$regex": regexWord, "$options": "i"}},
		},
		}
	} else {
		filter = bson.M{"date": primitive.NewDateTimeFromTime(dateSearch)}
	}
	cur, err := s.mgdb.Collection("pack").Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var packReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Value       float64   `bson:"value"`
		ProdType    string    `bson:"prodtype"`
		Itemcode    string    `bson:"itemcode"`
		ItemType    string    `bson:"itemtype"`
		Part        string    `bson:"part"`
		Factory     string    `bson:"factory"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &packReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/pack/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"packReports": packReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/pack/admin/reportdatefilter
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spka_reportdatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("packingFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("packingToDate"))

	filter := bson.M{"$and": bson.A{
		bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}},
		bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}},
	},
	}

	cur, err := s.mgdb.Collection("pack").Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var packReports []struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Value       float64   `bson:"value"`
		ProdType    string    `bson:"prodtype"`
		Itemcode    string    `bson:"itemcode"`
		ItemType    string    `bson:"itemtype"`
		Part        string    `bson:"part"`
		Factory     string    `bson:"factory"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &packReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/pack/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"packReports": packReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/pack/admin/deletereport/:reportid - delete a report on page admin of pack section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spka_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))

	deletedPackReport := s.mgdb.Collection("pack").FindOneAndDelete(context.Background(), bson.M{"_id": reportid})
	if deletedPackReport.Err() != nil {
		log.Println(deletedPackReport.Err())
		return
	}
	var packReport struct {
		ReportID string `bson:"_id"`
	}
	if err := deletedPackReport.Decode(&packReport); err != nil {
		log.Println(err)
	}
	refidObject, _ := primitive.ObjectIDFromHex(packReport.ReportID)
	// update production value
	result := s.mgdb.Collection("prodvalue").FindOneAndDelete(context.Background(), bson.M{"refid": refidObject})
	if result.Err() != nil {
		log.Println(result.Err())
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/pack/admin/updateform/:reportid
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spka_updateform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))
	result := s.mgdb.Collection("pack").FindOne(context.Background(), bson.M{"_id": reportid})
	if result.Err() != nil {
		log.Println(result.Err())
		return
	}
	var packReports struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Value       float64   `bson:"value"`
		ProdType    string    `bson:"prodtype"`
		Itemcode    string    `bson:"itemcode"`
		ItemType    string    `bson:"itemtype"`
		Part        string    `bson:"part"`
		Factory     string    `bson:"factory"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := result.Decode(&packReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/pack/admin/update_form.html")).Execute(w, map[string]interface{}{
		"packReports": packReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/pack/admin/updatereport/:reportid
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spka_updatereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))
	qty, _ := strconv.ParseFloat(r.FormValue("qty"), 64)
	value, _ := strconv.ParseFloat(r.FormValue("value"), 64)

	result := s.mgdb.Collection("pack").FindOneAndUpdate(context.Background(), bson.M{"_id": reportid}, bson.M{"$set": bson.M{"qty": qty, "value": value}})
	if result.Err() != nil {
		log.Println(result.Err())
		return
	}
	var packReports struct {
		ReportId    string    `bson:"_id"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		Value       float64   `bson:"value"`
		ProdType    string    `bson:"prodtype"`
		Itemcode    string    `bson:"itemcode"`
		ItemType    string    `bson:"itemtype"`
		Part        string    `bson:"part"`
		Factory     string    `bson:"factory"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := result.Decode(&packReports); err != nil {
		log.Println(err)
	}
	packReports.Qty = qty
	packReports.Value = value

	// update production value
	_, err := s.mgdb.Collection("prodvalue").UpdateOne(context.Background(), bson.M{"refid": reportid}, bson.M{"$set": bson.M{"qty": qty, "value": value}})
	if err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/pack/admin/updated_tr.html")).Execute(w, map[string]interface{}{
		"packReports": packReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/cutting/entry/woodrecoveryentry - get page entry of wood recovery of cutting section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sc_woodrecoveryentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/cutting/entry/woodrecovery.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/cutting/entry/woodrecoveryentry - get form
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sce_wr_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/wr_form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/cutting/entry/wr_sendentry - post form
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sce_wr_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	rate, _ := strconv.ParseFloat(r.FormValue("rate"), 64)
	prodtype := r.FormValue("prodtype")
	if r.FormValue("prodtype") == "" || r.FormValue("rate") == "" {
		template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/wr_form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	_, err := s.mgdb.Collection("woodrecovery").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "prodtype": prodtype, "rate": rate, "createdat": primitive.NewDateTimeFromTime(time.Now()), "reporter": username,
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/wr_form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/sections/cutting/entry/wr_form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/panelcnc/overview - get page overview of panelcnc
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) spc_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/panelcnc/overview/overview.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////////////
// /sections/panelcnc/overview/loadreport - load report table of page overview of panelcnc
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) spco_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("panelcnc").Aggregate(context.Background(), mongo.Pipeline{
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}},
			"startat":   bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$startat"}},
			"endat":     bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$endat"}},
			"createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var panelcncReports []struct {
		ReportId    string  `bson:"_id"`
		Machine     string  `bson:"machine"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		StartAt     string  `bson:"startat"`
		EndAt       string  `bson:"endat"`
		Hours       float64 `bson:"hours"`
		Type        string  `bson:"type"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &panelcncReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/panelcnc/overview/report.html")).Execute(w, map[string]interface{}{
		"panelcncReports": panelcncReports,
		"numberOfReports": len(panelcncReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/panelcnc/admin/searchreport - search reports on page admin of panelcnc section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spco_reportsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportsearch := r.FormValue("reportsearch")
	regexWord := ".*" + reportsearch + ".*"
	searchFilter := r.FormValue("searchFilter")

	cur, err := s.mgdb.Collection("panelcnc").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{searchFilter: bson.M{"$regex": regexWord, "$options": "i"}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}},
			"startat":   bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$date", "timezone": "Asia/Bangkok"}},
			"endat":     bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$endat", "timezone": "Asia/Bangkok"}},
			"createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var panelcncReports []struct {
		ReportId    string  `bson:"_id"`
		Machine     string  `bson:"machine"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		StartAt     string  `bson:"startat"`
		EndAt       string  `bson:"endat"`
		Hours       float64 `bson:"hours"`
		Type        string  `bson:"type"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &panelcncReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/panelcnc/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"panelcncReports": panelcncReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/panelcnc/overview/reportdatefilter
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spco_reportdatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("panelcncFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("panelcncToDate"))

	cur, err := s.mgdb.Collection("panelcnc").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}},
			"startat":   bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$date", "timezone": "Asia/Bangkok"}},
			"endat":     bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$endat", "timezone": "Asia/Bangkok"}},
			"createdat": bson.M{"$dateToString": bson.M{"format": "%H:%M ngày %d/%m", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var panelcncReports []struct {
		ReportId    string  `bson:"_id"`
		Machine     string  `bson:"machine"`
		Date        string  `bson:"date"`
		Qty         float64 `bson:"qty"`
		StartAt     string  `bson:"startat"`
		EndAt       string  `bson:"endat"`
		Hours       float64 `bson:"hours"`
		Type        string  `bson:"type"`
		Reporter    string  `bson:"reporter"`
		CreatedDate string  `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &panelcncReports); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/sections/panelcnc/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"panelcncReports": panelcncReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/panelcnc/entry - load page entry of panelcnc section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spc_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/panelcnc/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/panelcnc/entry/loadform - load form of page entry of panelcnc section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spc_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/panelcnc/entry/form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/panelcnc/entry/sendentry - post form of page entry of panelcnc section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spc_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	qty, _ := strconv.Atoi(r.FormValue("qty"))
	operator := r.FormValue("operator")
	paneltype := r.FormValue("type")
	machine := r.FormValue("machine")
	var start, end time.Time
	var hours float64
	now := time.Now()
	switch r.FormValue("timerange") {
	case "6h - 8h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.Local)
		hours = 2
	case "8h - 10h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.Local)
	case "10h - 11h30":
		start = time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 11, 30, 0, 0, time.Local)
	case "12h15 - 14h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 12, 15, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, time.Local)
	case "14h - 16h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, time.Local)
	case "16h30 - 18h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 16, 30, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.Local)
	case "18h - 20h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, time.Local)
	case "20h - 22h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, time.Local)
	case "22h30 - 0h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 22, 30, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	case "0h - 2h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, time.Local)
	case "2h45 - 4h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 2, 45, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, time.Local)
	case "4h - 6h":
		start = time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, time.Local)
		end = time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, time.Local)
	case "other":
		start, _ = time.ParseInLocation("2006-01-02T15:04", r.FormValue("start"), time.Local)
		end, _ = time.ParseInLocation("2006-01-02T15:04", r.FormValue("end"), time.Local)
	}
	hours = math.Round(end.Sub(start).Hours()*10) / 10
	date, _ := time.Parse("2006-01-02", start.Format("2006-01-02"))

	if machine == "" || r.FormValue("qty") == "" || hours <= 0 || r.FormValue("timerange") == "" {
		template.Must(template.ParseFiles("templates/pages/sections/panelcnc/entry/form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu hoặc sai, vui lòng nhập lại.",
		})
		return
	}
	_, err := s.mgdb.Collection("panelcnc").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "startat": primitive.NewDateTimeFromTime(start), "endat": primitive.NewDateTimeFromTime(end),
		"qty": qty, "createdat": primitive.NewDateTimeFromTime(time.Now()), "reporter": username,
		"machine": machine, "operator": operator, "type": paneltype, "hours": hours,
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/panelcnc/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/sections/panelcnc/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// ///////////////////////////////////////////////////////////////////////
// /sections/panelcnc/admin - get page admin of panelcnc section
// ///////////////////////////////////////////////////////////////////////
func (s *Server) spc_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/panelcnc/admin/admin.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////
// /sections/veneer/admin/loadreport - load report area on veneer admin page
// ///////////////////////////////////////////////////////////////////////
func (s *Server) spca_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("panelcnc").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdat": -1}).SetLimit(50))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var panelcncReports []struct {
		ReportId    string    `bson:"_id"`
		Machine     string    `bson:"machine"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		StartAt     time.Time `bson:"startat"`
		EndAt       time.Time `bson:"endat"`
		Hours       float64   `bson:"hours"`
		Type        string    `bson:"type"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err := cur.All(context.Background(), &panelcncReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/panelcnc/admin/report.html")).Execute(w, map[string]interface{}{
		"panelcncReports": panelcncReports,
		"numberOfReports": len(panelcncReports),
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/panelcnc/admin/searchreport - search reports on page admin of panelcnc section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spca_searchreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchWord := r.FormValue("reportSearch")
	regexWord := ".*" + searchWord + ".*"
	dateSearch, err := time.Parse("2006-01-02", searchWord)

	var filter bson.M
	if err != nil {
		filter = bson.M{"$or": bson.A{
			bson.M{"machine": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"type": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"reporter": bson.M{"$regex": regexWord, "$options": "i"}},
		},
		}
	} else {
		filter = bson.M{"date": primitive.NewDateTimeFromTime(dateSearch)}
	}
	cur, err := s.mgdb.Collection("panelcnc").Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var panelcncReports []struct {
		ReportId    string    `bson:"_id"`
		Machine     string    `bson:"machine"`
		Date        time.Time `bson:"date"`
		Qty         float64   `bson:"qty"`
		StartAt     time.Time `bson:"startat"`
		EndAt       time.Time `bson:"endat"`
		Hours       float64   `bson:"hours"`
		Type        string    `bson:"type"`
		Reporter    string    `bson:"reporter"`
		CreatedDate time.Time `bson:"createdat"`
	}
	if err = cur.All(context.Background(), &panelcncReports); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/sections/panelcnc/admin/report_tbody.html")).Execute(w, map[string]interface{}{
		"panelcncReports": panelcncReports,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/panelcnc/admin/deletereport/:reportid - delete a report on page admin of panelcnc section
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) spca_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reportid, _ := primitive.ObjectIDFromHex(ps.ByName("reportid"))

	_, err := s.mgdb.Collection("panelcnc").DeleteOne(context.Background(), bson.M{"_id": reportid})
	if err != nil {
		log.Println(err)
		return
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/outsource/entry
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sos_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/outsource/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/outsource/entry/loadform
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sose_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/sections/outsource/entry/form.html")).Execute(w, nil)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /sections/outsource/entry/sendentry
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sose_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameToken, _ := r.Cookie("username")
	username := usernameToken.Value
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))
	value, _ := strconv.ParseFloat(r.FormValue("value"), 64)
	factory := r.FormValue("factory")
	item := r.FormValue("item")
	qty, _ := strconv.Atoi(r.FormValue("qty"))

	if r.FormValue("value") == "" || r.FormValue("factory") == "" {
		template.Must(template.ParseFiles("templates/pages/sections/outsource/entry/form.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thông tin bị thiếu, vui lòng nhập lại.",
		})
		return
	}
	insertedResult, err := s.mgdb.Collection("outsource").InsertOne(context.Background(), bson.M{
		"date": primitive.NewDateTimeFromTime(date), "item": item, "factory": factory, "qty": qty, "value": value, "reporter": username, "createdat": primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/outsource/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}

	//create a report for production value collection
	_, err = s.mgdb.Collection("prodvalue").InsertOne(context.Background(), bson.M{
		"date":    primitive.NewDateTimeFromTime(date),
		"factory": factory, "prodtype": "outsource", "item": item, "qty": qty, "value": value, "reporter": username, "createdat": primitive.NewDateTimeFromTime(time.Now()),
		"from": "outsource", "refid": insertedResult.InsertedID,
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/sections/outsource/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối cơ sở dữ liệu thất bại, vui lòng nhập lại hoặc báo admin.",
		})
		return
	}

	template.Must(template.ParseFiles("templates/pages/sections/outsource/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /6s/overview - get page overview of 6S
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) s_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /6s/entry - get page entry of 6S
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) s6_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /6s/entry - send fast entry of 6S
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) s6_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /6s/admin - get admin page of 6S
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) s6_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /sections/packing/overview - get overview page of packing
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sp_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /sections/packing/entry - get entry page of packing
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sp_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func (s *Server) sp_loadentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func (s *Server) sp_mobystatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /sections/packing/entry/mosearch - search mo on packing entry page
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sp_mosearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////
// /sections/packing/entry/itemparts/:mo/:itemid - get form input when choose item
// ////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sp_itemparts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////
// /sections/packing/entry/itempart - chỉ nhập số lượng để khởi tạo part
// ////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sp_itempart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////
// /sections/packing/entry/initparts - initialize parts of item in mo
// ////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sp_initparts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////
// /sections/packing/entry/maxpartqtyinput - get max quantity of parts of item
// ////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sp_getinputmax(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /sections/packing/sendentry - create packing report, update motracking, check and create production value report
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sp_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /sections/packing/admin - get admin page of packing
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) sp_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// bản tạm cho packing
func (s *Server) sp_entrytmp(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/sections/packing/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// end bản tạm cho packing

// ////////////////////////////////////////////////////////////////////////////////////////////
// /production/overview - get page overview of production
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) p_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/production/overview/overview.html", "templates/shared/navbar.html")).Execute(w, nil)
}

// ///////////////////////////////////////////////////////////////////////////////
// /production/overview/loadprodtype - load chart prodtype of page overview of Production value
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) po_loadprodtype(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$expr": bson.M{"$eq": bson.A{bson.M{"$month": "$date"}, int(time.Now().Month())}}}}},
		{{"$group", bson.M{"_id": "$prodtype", "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.M{"value": -1}}},
		{{"$set", bson.M{"name": "$_id"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var prodtypeChartData []struct {
		Name  string  `bson:"name" json:"name"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &prodtypeChartData); err != nil {
		log.Println(err)
	}
	cur, err = s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$expr": bson.M{"$eq": bson.A{bson.M{"$month": "$date"}, int(time.Now().Month())}}}}},
		{{"$sort", bson.D{{"createdat", -1}, {"date", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d %H:%M", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}

	var rawData []struct {
		Date      string `bson:"date" json:"date"`
		CreatedAt string `bson:"createdat" json:"createdat"`
	}
	if err := cur.All(context.Background(), &rawData); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/production/overview/prodtype.html")).Execute(w, map[string]interface{}{
		"prodtypeChartData": prodtypeChartData,
		"rawData":           rawData,
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /production/overview/loadsummary - load summary table of page overview of Production value
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) po_loadsummary(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$expr": bson.M{"$eq": bson.A{bson.M{"$month": "$date"}, int(time.Now().Month())}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}, "qty": bson.M{"$sum": "$qty"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.prodtype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$_id.date"}}, "prodtype": "$_id.prodtype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var data []struct {
		Date     string  `bson:"date" json:"date"`
		Prodtype string  `bson:"prodtype" json:"prodtype"`
		Value    float64 `bson:"value" json:"value"`
		Qty      int     `bson:"qty" json:"qty"`
	}
	if err = cur.All(context.Background(), &data); err != nil {
		log.Println(err)
	}
	var mtdv, rhmtdv, brandmtdv, whitemtdv, outsourcemtdv float64
	var mtdp, rhmtdp, brandmtdp, whitemtdp, outsourcemtdp int
	var dates []string

	for _, i := range data {
		mtdv += i.Value
		mtdp += i.Qty
		switch i.Prodtype {
		case "brand":
			brandmtdv += i.Value
			brandmtdp += i.Qty
		case "rh":
			rhmtdv += i.Value
			rhmtdp += i.Qty
		case "white":
			whitemtdv += i.Value
			whitemtdp += i.Qty
		case "outsource":
			outsourcemtdv += i.Value
			outsourcemtdp += i.Qty
		}
		if !slices.Contains(dates, i.Date) {
			dates = append(dates, i.Date)
		}
	}

	pastdays := len(dates)

	var todayv, todaybrandv, todayrhv, todaywhitev, todayoutsourcev float64
	var todayp int
	todayv, todaybrandv, todayrhv, todayoutsourcev = 0, 0, 0, 0
	todayp = 0
	if pastdays >= 1 && time.Now().Add(7*time.Hour).Format("2006-01-02") == dates[len(dates)-1] {
		pastdays--

		for i := len(data) - 1; i > 0; i-- {
			if data[i].Date != dates[len(dates)-1] {
				break
			}
			todayv += data[i].Value
			todayp += data[i].Qty
			switch data[i].Prodtype {
			case "brand":
				todaybrandv += data[i].Value
			case "rh":
				todayrhv += data[i].Value
			case "white":
				todaywhitev += data[i].Value
			case "outsource":
				todayoutsourcev += data[i].Value
			}
		}
	}
	if pastdays == 0 {
		pastdays = 1
	}
	// cái náy dùng được, để sau này dùng
	var estdays int
	start := time.Now()
	end := time.Date(2024, time.Now().Month()+1, 1, 0, 0, 0, 0, time.Local)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if d.Weekday() != time.Sunday {
			estdays++
		}
	}

	// dùng tạm estdays tính tay này
	// estdays := 24 - pastdays

	p := message.NewPrinter(language.English)
	template.Must(template.ParseFiles("templates/pages/production/overview/summary.html")).Execute(w, map[string]interface{}{
		"month":         time.Now().Month().String(),
		"mtdv":          p.Sprintf("%.0f", mtdv),
		"mtdp":          p.Sprintf("%d", mtdp),
		"brandmtdv":     p.Sprintf("%.0f", brandmtdv),
		"brandmtdp":     p.Sprintf("%d", brandmtdp),
		"rhmtdv":        p.Sprintf("%.0f", rhmtdv),
		"rhmtdp":        p.Sprintf("%d", rhmtdp),
		"whitemtdv":     p.Sprintf("%.0f", whitemtdv),
		"whitemtdp":     p.Sprintf("%d", whitemtdp),
		"outsourcemtdv": p.Sprintf("%.0f", outsourcemtdv),
		"pastdays":      pastdays,
		"avgv":          p.Sprintf("%.0f", (mtdv-todayv)/float64(pastdays)),
		"avgp":          p.Sprintf("%d", mtdp/pastdays),
		"brandavgv":     p.Sprintf("%.0f", (brandmtdv-todaybrandv)/float64(pastdays)),
		"brandavgp":     p.Sprintf("%d", brandmtdp/pastdays),
		"rhavgv":        p.Sprintf("%.0f", (rhmtdv-todayrhv)/float64(pastdays)),
		"rhavgp":        p.Sprintf("%d", rhmtdp/pastdays),
		"whiteavgv":     p.Sprintf("%.0f", (whitemtdv-todaywhitev)/float64(pastdays)),
		"whiteavgp":     p.Sprintf("%d", whitemtdp/pastdays),
		"outsourceavgv": p.Sprintf("%.0f", (outsourcemtdv-todayoutsourcev)/float64(pastdays)),
		"estv":          p.Sprintf("%.0f", (mtdv-todayv)/float64(pastdays)*float64(estdays)+(mtdv-todayv)),
		"estbrandv":     p.Sprintf("%.0f", (brandmtdv-todaybrandv)/float64(pastdays)*float64(estdays)+(brandmtdv-todaybrandv)),
		"estrhv":        p.Sprintf("%.0f", (rhmtdv-todayrhv)/float64(pastdays)*float64(estdays)+(rhmtdv-todayrhv)),
		"estwhitev":     p.Sprintf("%.0f", (whitemtdv-todaywhitev)/float64(pastdays)*float64(estdays)+(whitemtdv-todaywhitev)),
		"estoutsourcev": p.Sprintf("%.0f", (outsourcemtdv-todayoutsourcev)/float64(pastdays)*float64(estdays)+(outsourcemtdv-todayoutsourcev)),
	})
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /production/overview/loadreport - load report table of page overview of production
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) po_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{}}},
		{{"$sort", bson.M{"createdat": -1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "at": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y %H:%M", "date": "$createdat"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var prodvalueData []struct {
		Id        string  `bson:"_id" json:"id"`
		Date      string  `bson:"date" json:"date"`
		Item      string  `bson:"item" json:"item"`
		ProdType  string  `bson:"prodtype" json:"prodtype"`
		ItemType  string  `bson:"itemtype" json:"itemtype"`
		Qty       int     `bson:"qty" json:"qty"`
		Value     float64 `bson:"value" json:"value"`
		From      string  `bson:"from" json:"from"`
		RefId     string  `bson:"refid" json:"refid"`
		Factory   string  `bson:"factory" json:"factory"`
		Reporter  string  `bson:"reporter" json:"reporter"`
		CreatedAt string  `bson:"at" json:"at"`
	}
	if err = cur.All(context.Background(), &prodvalueData); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/production/overview/report.html")).Execute(w, map[string]interface{}{
		"prodvalueData":   prodvalueData,
		"numberOfReports": len(prodvalueData),
	})
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /production/overview/reportfilter
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) po_reportfilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("reportFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("reportToDate"))

	cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$sort", bson.D{{"date", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}, "at": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y %H:%M", "date": "$createdat"}}}}},
	})
	if err != nil {
		log.Println("failed to access prodvalue at po_reportfilter")
	}
	defer cur.Close(context.Background())
	var prodvalueData []struct {
		Id        string  `bson:"_id" json:"id"`
		Date      string  `bson:"date" json:"date"`
		Item      string  `bson:"item" json:"item"`
		ProdType  string  `bson:"prodtype" json:"prodtype"`
		ItemType  string  `bson:"itemtype" json:"itemtype"`
		Qty       int     `bson:"qty" json:"qty"`
		Value     float64 `bson:"value" json:"value"`
		From      string  `bson:"from" json:"from"`
		RefId     string  `bson:"refid" json:"refid"`
		Factory   string  `bson:"factory" json:"factory"`
		Reporter  string  `bson:"reporter" json:"reporter"`
		CreatedAt string  `bson:"at" json:"at"`
	}
	if err = cur.All(context.Background(), &prodvalueData); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/production/overview/report_tbody.html")).Execute(w, map[string]interface{}{
		"prodvalueData": prodvalueData,
	})
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /production/overview/prodtypefilter
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) po_prodtypefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	start, _ := time.Parse("2006-01-02", r.FormValue("prodtypeFromDate"))
	end, _ := time.Parse("2006-01-02", r.FormValue("prodtypeToDate"))
	cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(start)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(end)}}}}}},
		{{"$group", bson.M{"_id": "$prodtype", "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.M{"value": -1}}},
		{{"$set", bson.M{"name": "$_id"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var prodtypeChartData []struct {
		Name  string  `bson:"name" json:"name"`
		Value float64 `bson:"value" json:"value"`
	}
	if err = cur.All(context.Background(), &prodtypeChartData); err != nil {
		log.Println(err)
	}
	cur, err = s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(start)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(end)}}}}}},
		{{"$sort", bson.D{{"date", -1}, {"createdat", -1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$date"}}, "createdat": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d %H:%M", "date": "$createdat", "timezone": "Asia/Bangkok"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	var rawData []struct {
		Date      string `bson:"date" json:"date"`
		CreatedAt string `bson:"createdat" json:"createdat"`
	}
	if err := cur.All(context.Background(), &rawData); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/production/overview/prodtypechart.html")).Execute(w, map[string]interface{}{
		"prodtypeChartData": prodtypeChartData,
		"rawData":           rawData,
	})
}

// ///////////////////////////////////////////////////////////////////////////////
// /production/overview/loadsummary - load summary table of page overview of Production value
// ///////////////////////////////////////////////////////////////////////////////
func (s *Server) po_summarydatefilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	month, _ := strconv.Atoi(r.FormValue("summarymonth"))
	cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$expr": bson.M{"$eq": bson.A{bson.M{"$month": "$date"}, month}}}}},
		{{"$group", bson.M{"_id": bson.M{"date": "$date", "prodtype": "$prodtype"}, "value": bson.M{"$sum": "$value"}, "qty": bson.M{"$sum": "$qty"}}}},
		{{"$sort", bson.D{{"_id.date", 1}, {"_id.prodtype", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$_id.date"}}, "prodtype": "$_id.prodtype"}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var data []struct {
		Date     string  `bson:"date" json:"date"`
		Prodtype string  `bson:"prodtype" json:"prodtype"`
		Value    float64 `bson:"value" json:"value"`
		Qty      int     `bson:"qty" json:"qty"`
	}
	if err = cur.All(context.Background(), &data); err != nil {
		log.Println(err)
	}
	if len(data) == 0 {
		template.Must(template.ParseFiles("templates/pages/production/overview/summary_tbody.html")).Execute(w, map[string]interface{}{})
		return
	}

	var mtdv, rhmtdv, brandmtdv, outsourcemtdv float64
	var mtdp, rhmtdp, brandmtdp, outsourcemtdp int
	var dates []string
	for _, i := range data {
		mtdv += i.Value
		mtdp += i.Qty
		switch i.Prodtype {
		case "brand":
			brandmtdv += i.Value
			brandmtdp += i.Qty
		case "rh":
			rhmtdv += i.Value
			rhmtdp += i.Qty
		case "outsource":
			outsourcemtdv += i.Value
			outsourcemtdp += i.Qty
		}
		if !slices.Contains(dates, i.Date) {
			dates = append(dates, i.Date)
		}
	}

	pastdays := len(dates)
	var todayv, todaybrandv, todayrhv, todayoutsourcev float64
	var todayp int
	if time.Now().Add(7*time.Hour).Format("2006-01-02") == dates[len(dates)-1] {
		pastdays--
		for i := len(data) - 1; i > 0; i-- {
			if data[i].Date != dates[len(dates)-1] {
				break
			}
			todayv += data[i].Value
			todayp += data[i].Qty
			switch data[i].Prodtype {
			case "brand":
				todaybrandv += data[i].Value
			case "rh":
				todayrhv += data[i].Value
			case "outsource":
				todayoutsourcev += data[i].Value
			}
		}
	}
	var estdays int
	if month != int(time.Now().Month()) {
		estdays = 0
	} else {
		start := time.Now()
		end := time.Date(2024, time.Now().Month()+1, 1, 0, 0, 0, 0, time.Local)
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			if d.Weekday() != time.Sunday {
				estdays++
			}
		}
	}

	p := message.NewPrinter(language.English)
	template.Must(template.ParseFiles("templates/pages/production/overview/summary_tbody.html")).Execute(w, map[string]interface{}{
		"mtdv":          p.Sprintf("%.0f", mtdv),
		"mtdp":          p.Sprintf("%d", mtdp),
		"brandmtdv":     p.Sprintf("%.0f", brandmtdv),
		"brandmtdp":     p.Sprintf("%d", brandmtdp),
		"rhmtdv":        p.Sprintf("%.0f", rhmtdv),
		"rhmtdp":        p.Sprintf("%d", rhmtdp),
		"outsourcemtdv": p.Sprintf("%.0f", outsourcemtdv),
		"pastdays":      pastdays,
		"avgv":          p.Sprintf("%.0f", mtdv/float64(pastdays)),
		"avgp":          p.Sprintf("%d", mtdp/pastdays),
		"brandavgv":     p.Sprintf("%.0f", brandmtdv/float64(pastdays)),
		"brandavgp":     p.Sprintf("%d", brandmtdp/pastdays),
		"rhavgv":        p.Sprintf("%.0f", rhmtdv/float64(pastdays)),
		"rhavgp":        p.Sprintf("%d", rhmtdp/pastdays),
		"outsourceavgv": p.Sprintf("%.0f", outsourcemtdv/float64(pastdays)),
		"estv":          p.Sprintf("%.0f", (mtdv-todayv)/float64(pastdays)*float64(estdays)+(mtdv-todayv)),
		"estbrandv":     p.Sprintf("%.0f", (brandmtdv-todaybrandv)/float64(pastdays)*float64(estdays)+(brandmtdv-todaybrandv)),
		"estrhv":        p.Sprintf("%.0f", (rhmtdv-todayrhv)/float64(pastdays)*float64(estdays)+(rhmtdv-todayrhv)),
		"estoutsourcev": p.Sprintf("%.0f", (outsourcemtdv-todayoutsourcev)/float64(pastdays)*float64(estdays)+(outsourcemtdv-todayoutsourcev)),
	})
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /target/entry - get page target entry
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) tg_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/target/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /target/entry/loadsectionentry
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) tge_loadsectionentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/target/entry/sectiontarget.html")).Execute(w, nil)
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /target/entry/loadreport
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) tge_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$sort", bson.D{{"date", -1}, {"name", 1}}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var targetData []struct {
		Id    string  `bson:"_id"`
		Date  string  `bson:"date"`
		Name  string  `bson:"name"`
		Value float64 `bson:"value"`
	}
	if err := cur.All(context.Background(), &targetData); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/target/entry/report.html")).Execute(w, map[string]interface{}{
		"targetData": targetData,
	})
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /target/entry/settarget - post settarget in page target entry
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) tge_settarget(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	targetname := r.FormValue("targetname")
	targetstart, _ := time.Parse("2006-01-02", r.FormValue("targetstart"))
	targetend, _ := time.Parse("2006-01-02", r.FormValue("targetend"))
	weekdays := strings.Fields(r.FormValue("weekdays"))
	target, _ := strconv.Atoi(r.FormValue("target"))

	if targetname == "" || r.FormValue("target") == "" || r.FormValue("weekdays") == "" {
		template.Must(template.ParseFiles("templates/pages/target/entry/sectiontarget.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thiếu thông tin, vui lòng nhập lại.",
		})
		return
	}

	var intWeekDays []int
	for _, d := range weekdays {
		t, _ := strconv.Atoi(d)
		intWeekDays = append(intWeekDays, t)
	}

	// var bdoc []interface{}
	for tmpdate := targetstart; tmpdate.Before(targetend.AddDate(0, 0, 1)); tmpdate = tmpdate.AddDate(0, 0, 1) {
		if slices.Contains(intWeekDays, int(tmpdate.Weekday())) {
			// b := bson.M{
			// 	"name": targetname, "date": primitive.NewDateTimeFromTime(tmpdate), "value": target,
			// }
			// bdoc = append(bdoc, b)
			_, err := s.mgdb.Collection("target").UpdateOne(context.Background(), bson.M{"name": targetname, "date": primitive.NewDateTimeFromTime(tmpdate)}, bson.M{
				"$set": bson.M{"value": target},
			}, options.Update().SetUpsert(true))
			if err != nil {
				log.Println(err)
				template.Must(template.ParseFiles("templates/pages/target/entry/sectiontarget.html")).Execute(w, map[string]interface{}{
					"showErrDialog": true,
					"msgDialog":     "Cập nhật thất bại, vui lòng nhập lại.",
				})
				return
			}
		}
	}

	// _, err := s.mgdb.Collection("target").InsertMany(context.Background(), bdoc, options.InsertMany())
	// if err != nil {
	// 	log.Println(err)
	// 	template.Must(template.ParseFiles("templates/pages/target/entry/sectiontarget.html")).Execute(w, map[string]interface{}{
	// 		"showErrDialog": true,
	// 		"msgDialog":     "Cập nhật thất bại, vui lòng nhập lại.",
	// 	})
	// 	return
	// }

	template.Must(template.ParseFiles("templates/pages/target/entry/sectiontarget.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Đã đặt target thành công",
	})
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /target/entry/loadplanworkdays
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) tge_loadplanworkdays(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/target/entry/planworkdays.html")).Execute(w, nil)
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /target/entry/setworkdays
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) tge_setworkdays(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	month, _ := strconv.Atoi(r.FormValue("month"))
	workdays, _ := strconv.Atoi(r.FormValue("workdays"))

	if r.FormValue("month") == "" || r.FormValue("workdays") == "" {
		template.Must(template.ParseFiles("templates/pages/target/entry/planworkdays.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thiếu thông tin, vui lòng nhập lại.",
		})
		return
	}
	date := time.Date(time.Now().Year(), time.Month(month), 15, 0, 0, 0, 0, time.Local)

	_, err := s.mgdb.Collection("target").InsertOne(context.Background(), bson.M{
		"name": "plan work days", "date": primitive.NewDateTimeFromTime(date), "value": workdays,
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/target/entry/planworkdays.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Cập nhật thất bại, vui lòng nhập lại.",
		})
		return
	}

	template.Must(template.ParseFiles("templates/pages/target/entry/planworkdays.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Đã đặt số ngày dự kiến thành công",
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /target/entry/search
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) tge_search(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchRegex := ".*" + r.FormValue("targetSearch") + ".*"
	cur, err := s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"name": bson.M{"$regex": searchRegex, "$options": "i"}}}},
		{{"$sort", bson.M{"date": -1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var targetData []struct {
		Id    string  `bson:"_id"`
		Date  string  `bson:"date"`
		Name  string  `bson:"name"`
		Value float64 `bson:"value"`
	}
	if err := cur.All(context.Background(), &targetData); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/target/entry/target_tbody.html")).Execute(w, map[string]interface{}{
		"targetData": targetData,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /target/entry/filterbydate
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) tge_filterbydate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	occurdate, _ := time.Parse("2006-01-02", r.FormValue("occurdate"))

	cur, err := s.mgdb.Collection("target").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"date": primitive.NewDateTimeFromTime(occurdate)}}},
		{{"$sort", bson.M{"name": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var targetData []struct {
		Id    string  `bson:"_id"`
		Date  string  `bson:"date"`
		Name  string  `bson:"name"`
		Value float64 `bson:"value"`
	}
	if err := cur.All(context.Background(), &targetData); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/target/entry/target_tbody.html")).Execute(w, map[string]interface{}{
		"targetData": targetData,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /target/entry/deletereport/:id
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) tge_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := primitive.ObjectIDFromHex(ps.ByName("id"))

	_, err := s.mgdb.Collection("target").DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		log.Println(err)
		return
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /manhr/admin - get page manhr admin
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) m_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/manhr/admin/admin.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /manhr/admin/loadentry - load entry section
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) ma_loadentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/manhr/admin/manhrentry.html")).Execute(w, nil)
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /manhr/admin/loadreport - load manhr table section
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) ma_loadreport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
		{{"$sort", bson.D{{"date", -1}, {"section", 1}}}},
		{{"$limit", 50}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var manhrData []struct {
		Id      string  `bson:"_id"`
		Date    string  `bson:"date"`
		Section string  `bson:"section"`
		Hc      int     `bson:"hc"`
		Workhr  float64 `bson:"workhr"`
	}
	if err := cur.All(context.Background(), &manhrData); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/manhr/admin/report.html")).Execute(w, map[string]interface{}{
		"manhrData":       manhrData,
		"numberOfReports": len(manhrData),
	})
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /manhr/admin/sendentry - send entry of manhr
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) ma_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	section := r.FormValue("section")
	date, _ := time.Parse("2006-01-02", r.FormValue("occurdate"))
	hc, _ := strconv.Atoi(r.FormValue("hc"))
	workhr, _ := strconv.ParseFloat(r.FormValue("workhr"), 64)
	if section == "" || hc == 0 || workhr == 0 {
		template.Must(template.ParseFiles("templates/pages/manhr/admin/manhrentry.html")).Execute(w, map[string]interface{}{
			"showMissingDialog": true,
			"msgDialog":         "Thiếu thông tin nhập liệu",
		})
		return
	}
	_, err := s.mgdb.Collection("manhr").InsertOne(context.Background(), bson.M{
		"section": section, "date": primitive.NewDateTimeFromTime(date), "hc": hc, "workhr": workhr,
	})
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/manhr/admin/manhrentry.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Record này có thể đã có rồi, check và update thay vì tạo mới.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/manhr/admin/manhrentry.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Cập nhật thành công",
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /manhr/admin/deletereport/:id
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) ma_deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := primitive.ObjectIDFromHex(ps.ByName("id"))

	_, err := s.mgdb.Collection("manhr").DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		log.Println(err)
		return
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /manhr/admin/updateform/:id
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) ma_updateform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := primitive.ObjectIDFromHex(ps.ByName("id"))

	result := s.mgdb.Collection("manhr").FindOne(context.Background(), bson.M{"_id": id})
	if result.Err() != nil {
		log.Println(result.Err())
		return
	}
	var manhrData struct {
		Id      string    `bson:"_id"`
		Date    time.Time `bson:"date"`
		Section string    `bson:"section"`
		Hc      int       `bson:"hc"`
		Workhr  float64   `bson:"workhr"`
	}
	if err := result.Decode(&manhrData); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/manhr/admin/update_form.html")).Execute(w, map[string]interface{}{
		"manhrData": manhrData,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /manhr/admin/updatereport/:id
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) ma_updatereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := primitive.ObjectIDFromHex(ps.ByName("id"))
	hc, _ := strconv.Atoi(r.FormValue("hc"))
	workhr, _ := strconv.ParseFloat(r.FormValue("workhr"), 64)

	result := s.mgdb.Collection("manhr").FindOneAndUpdate(context.Background(), bson.M{"_id": id}, bson.M{"$set": bson.M{"hc": hc, "workhr": workhr}})
	if result.Err() != nil {
		log.Println(result.Err())
		return
	}
	var manhrData struct {
		Id      string    `bson:"_id"`
		Date    time.Time `bson:"date"`
		Section string    `bson:"section"`
		Hc      int       `bson:"hc"`
		Workhr  float64   `bson:"workhr"`
		DateStr string
	}
	if err := result.Decode(&manhrData); err != nil {
		log.Println(err)
	}
	manhrData.Hc = hc
	manhrData.Workhr = workhr
	manhrData.DateStr = manhrData.Date.Format("02-01-2006")

	template.Must(template.ParseFiles("templates/pages/manhr/admin/updated_tr.html")).Execute(w, map[string]interface{}{
		"manhrData": manhrData,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /manhr/admin/search
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) ma_search(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchRegex := ".*" + r.FormValue("manhrSearch") + ".*"
	cur, err := s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"section": bson.M{"$regex": searchRegex, "$options": "i"}}}},
		{{"$sort", bson.M{"date": -1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var manhrData []struct {
		Id      string  `bson:"_id"`
		Date    string  `bson:"date"`
		Section string  `bson:"section"`
		Hc      int     `bson:"hc"`
		Workhr  float64 `bson:"workhr"`
	}
	if err := cur.All(context.Background(), &manhrData); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/manhr/admin/manhr_tbody.html")).Execute(w, map[string]interface{}{
		"manhrData": manhrData,
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /manhr/admin/filterbydate
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) ma_filterbydate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	occurdate, _ := time.Parse("2006-01-02", r.FormValue("occurdate"))

	cur, err := s.mgdb.Collection("manhr").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"date": primitive.NewDateTimeFromTime(occurdate)}}},
		{{"$sort", bson.M{"section": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d-%m-%Y", "date": "$date"}}}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())
	var manhrData []struct {
		Id      string  `bson:"_id"`
		Date    string  `bson:"date"`
		Section string  `bson:"section"`
		Hc      int     `bson:"hc"`
		Workhr  float64 `bson:"workhr"`
	}
	if err := cur.All(context.Background(), &manhrData); err != nil {
		log.Println(err)
	}
	template.Must(template.ParseFiles("templates/pages/manhr/admin/manhr_tbody.html")).Execute(w, map[string]interface{}{
		"manhrData": manhrData,
	})
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// router.POST("/manhr/entry/sendtotalmanhr", s.me_sendtotalmanhr)
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) me_sendtotalmanhr(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	date, _ := time.Parse("2006-01-02", r.FormValue("vopdate"))
	totalmanhr, _ := strconv.ParseFloat(r.FormValue("totalmanhr"), 64)

	if r.FormValue("totalmanhr") == "" {
		w.Write([]byte("thiếu thông tin"))
		return
	}
	_, err := s.mgdb.Collection("vopmanhr").UpdateOne(context.Background(), bson.M{"date": primitive.NewDateTimeFromTime(date)}, bson.M{
		"$set": bson.M{"manhr": totalmanhr},
	}, options.Update().SetUpsert(true))
	if err != nil {
		log.Println(err)
	}

	// load chart
	fromdate, _ := time.Parse("2006-01-02", r.FormValue("vopFromDate"))
	todate, _ := time.Parse("2006-01-02", r.FormValue("vopToDate"))

	cur, err := s.mgdb.Collection("prodvalue").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$group", bson.M{"_id": "$date", "value": bson.M{"$sum": "$value"}}}},
		{{"$sort", bson.M{"_id": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$_id"}}}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var productiondata []struct {
		Date  string  `json:"date"`
		Value float64 `json:"value"`
	}

	if err := cur.All(context.Background(), &productiondata); err != nil {
		log.Println(err)
	}

	cur, err = s.mgdb.Collection("vopmanhr").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"date": bson.M{"$gte": primitive.NewDateTimeFromTime(fromdate)}}, bson.M{"date": bson.M{"$lte": primitive.NewDateTimeFromTime(todate)}}}}}},
		{{"$sort", bson.M{"date": 1}}},
		{{"$set", bson.M{"date": bson.M{"$dateToString": bson.M{"format": "%d %b", "date": "$date"}}}}},
		{{"$unset", "_id"}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var manhrdata []struct {
		Date  string  `bson:"date" json:"date"`
		Manhr float64 `bson:"manhr" json:"manhr"`
	}

	if err := cur.All(context.Background(), &manhrdata); err != nil {
		log.Println(err)
	}

	template.Must(template.ParseFiles("templates/pages/dashboard/productionvop_genchart.html")).Execute(w, map[string]interface{}{
		"productiondata": productiondata,
		"manhrdata":      manhrdata,
	})
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /downtime/entry - copy paste report for downtime
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) dt_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /downtime//entry/loadform - load form of report for downtime
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) dte_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /downtime/sendentry - post report for downtime
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) dte_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/admin", s.c_admin)
func (s *Server) c_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/admin/loadusingtimeform", s.ca_loadusingtimeform)
func (s *Server) ca_loadusingtimeform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/admin/loadinspectionform", s.ca_loadinspectionform)
func (s *Server) ca_loadinspectionform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/colormixing/admin/addinspection", s.ca_addinspection)

// router.POST("/mixingcolor/getusingstart", s.getusingstart)
func (s *Server) getusingstart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/loaddeliveryentry
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) loaddeliveryentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/senddeliveryentry
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) senddeliveryentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/batchentry
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) mc_batchentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// // cur, err := s.mgdb.Collection("colorpanel").Find(context.Background(), bson.M{})
	// // if err != nil {
	// // 	log.Println(err)
	// // }
	// // defer cur.Close(context.Background())
	// // var colorpanelData []struct {
	// // 	Code     string `bson:"code"`
	// // 	Color    string `bson:"color"`
	// // 	Brand    string `bson:"brand"`
	// // 	Supplier string `bson:"supplier"`
	// // }
	// // if err := cur.All(context.Background(), &colorpanelData); err != nil {
	// // 	log.Println(err)
	// // }

	// template.Must(template.ParseFiles(
	// 	"templates/pages/mixingcolor/entry/batchentry.html",
	// 	"templates/shared/navbar.html",
	// )).Execute(w, nil)
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/entry/loadbatchform
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) mce_loadbatchform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// cur, err := fasdfass.mgdb.Collection("colorpanel").Find(context.Background(), bson.M{})
	// if err != nisdfasdl {
	// 	log.Println(err)
	// }
	// defer cur.Close(contedfasdfaxt.Background())
	// var colorpanelData []dfasstruct {
	// 	Code     string `bson:"code"`
	// 	Color    stdfasdring `bson:"csdfasdfasolor"`
	// 	Brand    string `bson:"brand"`
	// 	Supplier string `bson:"supplier"`
	// }fas
	// if err := cur.Afasdfasll(context.Background(), &colorpanelData); err != nil {
	// 	log.Println(edfasdrr)
	// }

	// template.Must(teasdfmplate.ParseFiles(
	// 	"templates/pagedfs/mixingcolor/entry/batchform.html",
	// )).Execute(w, masdfasp[string]interface{}{
	// 	"colorpanelData"df: colorpanelData,
	// })
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/entry/sendbatchentry
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) mce_sendbatchentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/admin/loadbatchentry", s.ca_loadbatchentry)
func (s *Server) ca_loadbatchentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/admin/loadpanelentry", s.ca_loadpanelentry)
func (s *Server) ca_loadpanelentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

// router.POST("/colormixing/admin/sendpanelentry", s.ca_sendpanelentry)

// router.POST("/colormixing/admin/sendbatchentry", s.ca_sendmixingentry)

// router.GET("/colormixing/admin/loadmixingbatch", s.ca_loadmixingbatch)
func (s *Server) ca_loadmixingbatch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/colormixing/admin/searchbatch", s.ca_searchbatch)
func (s *Server) ca_searchbatch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/mixingfilter
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) mixingfilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/admin/batchupdateform/:batchno", s.ca_batchupdateform)
func (s *Server) ca_batchupdateform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/admin/loadauditentry", s.ca_loadauditentry)
func (s *Server) ca_loadauditentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/admin/failaudit/:id", s.ca_failaudit)
func (s *Server) ca_failaudit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/colormixing/admin/passaudit/:id", s.ca_passaduti)
func (s *Server) ca_passaduti(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/mixingreports/:batchno
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) mixingreports(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// mixingcolor/deletereports
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) deletereport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.DELETE("/colormixing/admin/deletemixing/:batchno", s.ca_deletemixing)
func (s *Server) ca_deletemixing(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/addcolorform
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) addcolorform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/mixingcolor/colorform_tr.html")).Execute(w, nil)
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/addcolor
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) addcolor(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/colorsearch
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) colorsearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /mixingcolor/colorfilter
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) colorfilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// operator := r.FormValue("operator")
	// color := r.FormValue("color")
	// code := r.FormValue("code")
	// brand := r.FormValue("brand")
	// supplier := r.FormValue("supplier")
	// classification := r.FormValue("classification")
	// sopno := r.FormValue("sopno")
	// status := r.FormValue("status")
	// mixingdate, _ := time.Parse("2006-01-02", r.FormValue("mixingdate"))

	// if r.FormValue("mixingdate") != "" {
	// 	filters = append(filters, bson.M{"$and": bson.A{bson.M{"mixingdate": bson.M{"$gte": primitive.NewDateTimeFromTime(mixingdate)}}, bson.M{"mixingdate": bson.M{"$lt": primitive.NewDateTimeFromTime(mixingdate.AddDate(0, 0, 1))}}}})
	// }

}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// router.DELETE("/colormixing/admin/deletepanel/:id", s.ca_deletepanel)
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) ca_deletepanel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/admin/panelupdateform/:id}", s.ca_panelupdateform)
func (s *Server) ca_panelupdateform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.PUT("/colormixing/admin/updatepanel/:id", s.ca_updatepanel)
func (s *Server) ca_updatepanel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/mixingcolor/usingentry", s.mc_usingreports)
func (s *Server) mc_usingreports(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// template.Must(template.ParseFiles(
	// 	"templates/pages/mixingcolor/entry/usingentry.html",
	// 	"templates/shared/navbar.html",
	// )).Execute(w, nil)
}

// router.GET("/mixingcolor/entry/loadusingform", s.mc_loadusingform)
func (s *Server) mc_loadusingform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	template.Must(template.ParseFiles(
		"templates/pages/mixingcolor/entry/usingform.html",
	)).Execute(w, nil)
}

// router.GET("/mixingcolor/entry/getupdateform", s.mc_getupdateform)
func (s *Server) mc_getupdateform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	switch r.URL.Query().Get("formtype") {
	case "usingtime":
		cur, err := s.mgdb.Collection("mixingbatch").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"issueddate": -1}))
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var batchData []struct {
			BatchNo string `bson:"batchno"`
		}
		if err := cur.All(context.Background(), &batchData); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/mixingcolor/entry/usingtimeform.html")).Execute(w, map[string]interface{}{
			"batchData": batchData,
		})
	case "usingitem":
		cur, err := s.mgdb.Collection("mixingbatch").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"issueddate": -1}))
		if err != nil {
			log.Println(err)
		}
		defer cur.Close(context.Background())
		var batchData []struct {
			BatchNo string `bson:"batchno"`
		}
		if err := cur.All(context.Background(), &batchData); err != nil {
			log.Println(err)
		}
		template.Must(template.ParseFiles("templates/pages/mixingcolor/entry/usingitemform.html")).Execute(w, map[string]interface{}{
			"batchData": batchData,
		})

	case "createcolor":
		template.Must(template.ParseFiles("templates/pages/mixingcolor/entry/createcolorform.html")).Execute(w, nil)
	case "fastbatch":
		template.Must(template.ParseFiles("templates/pages/mixingcolor/entry/fastbatchform.html")).Execute(w, nil)
	}

}

// router.GET("/mixingcolor/entry/updateusingtime", s.mc_updateusingtime)
func (s *Server) mc_updateusingtime(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/mixingcolor/entry/updateusingitem", s.updateusingitem)
func (s *Server) mc_updateusingitem(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/mixingcolor/entry/searchcolorcode", s.mce_searchcolorcode)
func (s *Server) mce_searchcolorcode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/mixingcolor/entry/createfastbatch", s.mce_createfastbatch)
func (s *Server) mce_createfastbatch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/mixingcolor/colorentry", s.mc_colorentry)
func (s *Server) mc_colorentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/mixingcolor/entry/colorentry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// router.GET("/mixingcolor/entry/loadcolorform", s.mc_loadcolorform)
func (s *Server) mc_loadcolorform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/overview", s.c_overview)
func (s *Server) c_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/overview/loadbatch", s.co_loadbatch)
func (s *Server) co_loadbatch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cur, err := s.mgdb.Collection("mixingbatch").Aggregate(context.Background(), mongo.Pipeline{
		{{"$match", bson.M{"$and": bson.A{bson.M{"mixingdate": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -3))}}, bson.M{"mixingdate": bson.M{"$lte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, 1))}}}}}},
		{{"$sort", bson.D{{"mixingdate", -1}, {"batchno", 1}}}},
		{{"$set", bson.M{
			"mixingdate": bson.M{"$dateToString": bson.M{"format": "%H:%M %d-%m-%Y", "date": "$mixingdate"}},
			"issueddate": bson.M{"$dateToString": bson.M{"format": "%H:%M %d-%m-%Y", "date": "$issueddate"}},
			"startuse":   bson.M{"$dateToString": bson.M{"format": "%H:%M %d-%m-%Y", "date": "$startuse"}},
			"enduse":     bson.M{"$dateToString": bson.M{"format": "%H:%M %d-%m-%Y", "date": "$enduse"}},
		}}},
	})
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(context.Background())

	var mixingbatchData []models.BatchRecord_datestr
	if err := cur.All(context.Background(), &mixingbatchData); err != nil {
		log.Println(err)
	}

	var operatorMap = make(map[string]bool, len(mixingbatchData))
	var colorMap = make(map[string]bool, len(mixingbatchData))
	var codeMap = make(map[string]bool, len(mixingbatchData))
	var brandMap = make(map[string]bool, len(mixingbatchData))
	var supplierMap = make(map[string]bool, len(mixingbatchData))
	var classificationMap = make(map[string]bool, len(mixingbatchData))
	var sopnoMap = make(map[string]bool, len(mixingbatchData))
	var statusMap = make(map[string]bool, len(mixingbatchData))

	for _, v := range mixingbatchData {
		operatorMap[v.Operator] = true
		codeMap[v.Color.Code] = true
		colorMap[v.Color.Name] = true
		brandMap[v.Color.Brand] = true
		supplierMap[v.Color.Supplier] = true
		classificationMap[v.Classification] = true
		sopnoMap[v.SOPNo] = true
		statusMap[v.Status] = true
	}
	var operators = make([]string, 0, len(operatorMap))
	for k, _ := range operatorMap {
		operators = append(operators, k)
	}
	var colors = make([]string, 0, len(colorMap))
	for k, _ := range colorMap {
		colors = append(colors, k)
	}
	var codes = make([]string, 0, len(codeMap))
	for k, _ := range codeMap {
		codes = append(codes, k)
	}
	var brands = make([]string, 0, len(brandMap))
	for k, _ := range brandMap {
		brands = append(brands, k)
	}
	var suppliers = make([]string, 0, len(supplierMap))
	for k, _ := range supplierMap {
		suppliers = append(suppliers, k)
	}
	var classifications = make([]string, 0, len(classificationMap))
	for k, _ := range classificationMap {
		classifications = append(classifications, k)
	}
	var sopnos = make([]string, 0, len(sopnoMap))
	for k, _ := range sopnoMap {
		sopnos = append(sopnos, k)
	}
	var statuses = make([]string, 0, len(statusMap))
	for k, _ := range statusMap {
		statuses = append(statuses, k)
	}

	template.Must(template.ParseFiles("templates/pages/colormixing/overview/batch.html")).Execute(w, map[string]interface{}{
		"mixingbatchData": mixingbatchData,
		"operators":       operators,
		"colors":          colors,
		"codes":           codes,
		"brands":          brands,
		"suppliers":       suppliers,
		"classifications": classifications,
		"sopnos":          sopnos,
		"statuses":        statuses,
	})
}

// router.GET("/colormixing/overview/searchbatch", s.co_searchbatch)
func (s *Server) co_searchbatch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/colormixing/overview/changedisplay/:type/edit/false", s.co_changedisplay)
func (s *Server) co_changedisplay(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/colormixing/overview/searchcolor", s.co_searchcolor)
func (s *Server) co_searchcolor(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/colormixing/overview/filtercolor", s.co_filtercolor)
func (s *Server) co_filtercolor(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// router.POST("/colormixing/overview/createstandard", s.co_createstandard)
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) co_createstandard(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// router.DELETE("/colormixing/overview/deletestandard/:name", s.co_deletestandard)
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) co_deletestandard(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /quality/entry - copy paste report for quality
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) q_fastentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles(
		"templates/pages/quality/entry/entry.html",
		"templates/shared/navbar.html",
	)).Execute(w, nil)
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /quality/loadform - load form of report for quality
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) q_loadform(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	template.Must(template.ParseFiles("templates/pages/quality/entry/form.html")).Execute(w, nil)
}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /quality/sendentry - post report for quality
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) q_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	lines := strings.Split(strings.Trim(r.FormValue("list"), "\n"), "\n")
	date, _ := time.Parse("Jan 02, 2006", r.FormValue("occurdate"))

	var jsonStr = `[`
	for _, line := range lines {
		raw := strings.Fields(line)
		section := raw[0]
		checkedqty := raw[1]
		failedqty := "0"
		if len(raw) == 3 {
			failedqty = raw[2]
		}

		jsonStr += `{
			"date":"` + date.Format("2006-01-02") + `", 
			"section":"` + section + `", 
			"checkedqty":` + checkedqty + `,
			"failedqty":` + failedqty + `
			},`
	}
	jsonStr = jsonStr[:len(jsonStr)-1] + `]`

	var bdoc []interface{}
	err := bson.UnmarshalExtJSON([]byte(jsonStr), true, &bdoc)
	if err != nil {
		log.Print(err)
		template.Must(template.ParseFiles("templates/pages/quality/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Lỗi decode. Vui lòng liên hệ admin.",
		})
		return
	}
	_, err = s.mgdb.Collection("quality").InsertMany(context.Background(), bdoc)
	if err != nil {
		log.Println(err)
		template.Must(template.ParseFiles("templates/pages/quality/entry/form.html")).Execute(w, map[string]interface{}{
			"showErrDialog": true,
			"msgDialog":     "Kết nối database thất bại. Vui lòng liên hệ admin.",
		})
		return
	}
	template.Must(template.ParseFiles("templates/pages/quality/entry/form.html")).Execute(w, map[string]interface{}{
		"showSuccessDialog": true,
		"msgDialog":         "Gửi dữ liệu thành công.",
	})
}

// //////////////////////////////////////////////
// router.GET("/gnhh/overview", s.g_overview)
// //////////////////////////////////////////////
func (s *Server) g_overview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/gnhh/overview/loadchart", s.go_loadchart)
func (s *Server) go_loadchart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ///////////////////////////////////////////////////////////////////
// router.GET("/gnhh/overview/loadproducttree", s.go_loadproducttree)
// ///////////////////////////////////////////////////////////////////
func (s *Server) go_loadproducttree(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/gnhh/overview/loadtimeline", s.go_loadtimeline)
func (s *Server) go_loadtimeline(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/gnhh/overview/loaddetail", s.go_loaddetail)
func (s *Server) go_loaddetail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/gnhh/overview/updatetimeline", s.go_updatetimeline)
func (s *Server) go_updatetimeline(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/gnhh/overview/searchtimeline", s.go_searchtimeline)
func (s *Server) go_searchtimeline(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/gnhh/overview/filtertimeline", s.go_filtertimeline)
func (s *Server) go_filtertimeline(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/gnhh/overview/iteminfo", s.go_iteminfo)
func (s *Server) go_iteminfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/gnhh/overview/loadtree", s.go_loadtree)
func (s *Server) go_loadtree(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/gnhh/overview/getproductcodes", s.go_getproductcodes)
func (s *Server) go_getproductcodes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/gnhh/overview/mofilter", s.go_mofilter)
func (s *Server) go_mofilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/gnhh/overview/productfilter", s.go_productfilter)
func (s *Server) go_productfilter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/gnhh/overview/searchdetail", s.go_searchdetail)
func (s *Server) go_searchdetail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.GET("/gnhh/entry/import", s.ge_import)
func (s *Server) ge_import(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// router.POST("/gnhh/entry/importdata", s.ge_importdata)
func (s *Server) ge_importdata(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /safety/entry
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) s_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// ////////////////////////////////////////////////////////////////////////////////////////////
// /safety/sendentry
// ////////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) s_sendentry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// /////////////////////////////////////////////////////
// router.POST("/manpower/importfile", s.mp_importfile)
// /////////////////////////////////////////////////////
func (s *Server) mp_importfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// /////////////////////////////////////////////////////
// router.POST("/otattend/importfile", s.oa_importfile)
// //////////////////////////////////////////////////////
func (s *Server) oa_importfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// /////////////////////////////////////////////////////////////////////////////////////////
// /mo/entry - get entry page of mo
// /////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) mo_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// /////////////////////////////////////////////////////////////////////////////////////////
// /mo/entry - post entry page of mo
// /////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) mo_insertMoList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func (s *Server) mo_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// /////////////////////////////////////////////////////////////////////////////////////////
// /item/entry - get entry page of item
// /////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) i_entry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func (s *Server) i_importitemlist(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

// /////////////////////////////////////////////////////////////////////////////////////////
// /item/admin - get admin page of item
// /////////////////////////////////////////////////////////////////////////////////////////
func (s *Server) i_admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func (s *Server) i_addpart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func (s *Server) handleGetTest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func (s *Server) testload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}
