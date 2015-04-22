/* @flow */

var React = require("react");
window.React = React; // export for http://fb.me/react-devtools
var Backbone = require("backbone");
Backbone.$ = window.$;

var Accounts = require("./accounts.js");
var Analytics = require("./analytics.js");
var CPMappings = require("./cp_mappings.js");
var Dashboard = require("./dashboard.js");
var Doctors = require("./doctors.js");
var Drugs = require("./drugs.js");
var Guides = require("./guides.js");
var Pathways = require("./pathways.js");
var Perms = require("./permissions.js");
var Financial = require("./financial.js");
var FavoriteTreatmentPlan = require("./favorite_treatment_plan.js");
var Nav = require("../../libs/nav.js");
var Visit = require("./visit.js");

var AdminRouter = Backbone.Router.extend({
	routes : {
		"": function() {
			this.current = "dashboard";
			this.params = {};
		},
		"doctors": function() {
			this.current = "doctorSearch";
			this.params = {};
		},
		"doctors/:doctorID/:page": function(doctorID, page) {
			this.current = "doctor";
			this.params = {doctorID: doctorID, page: page};
		},
		"care_provider_mappings": function() {
			this.current = "careProviderMappings";
			this.params = {};
		},
		"guides/:page": function(page) {
			this.current = "guides";
			this.params = {page: page};
		},
		"guides/:page/:id": function(page, guideID) {
			this.current = "guides";
			this.params = {page: page, guideID: guideID};
		},
		"analytics/:page": function(page) {
			this.current = "analytics";
			this.params = {page: page};
		},
		"analytics/:page/:id": function(page, reportID) {
			this.current = "analytics";
			this.params = {page: page, reportID: reportID};
		},
		"accounts": function() {
			this.current = "accountsList";
			this.params = {};
		},
		"accounts/:accountID/:page": function(accountID, page) {
			this.current = "account";
			this.params = {accountID: accountID, page: page};
		},
		"drugs": function() {
			this.current = "drugs";
			this.params = {};
		},
		"pathways": function() {
			this.current = "pathways";
			this.params = {page: "list"};
		},
		"pathways/:page": function(page) {
			this.current = "pathways";
			this.params = {page: page};
		},
		"pathways/:page/:id": function(page, pathwayID) {
			this.current = "pathways";
			this.params = {page: page, pathwayID: pathwayID};
		},
		"financial": function() {
			this.current = "financial";
			this.params = {page: "incoming"}
		},
		"financial/:page": function(page) {
			this.current = "financial";
			this.params = {page: page};
		},
		"treatment_plan/favorite/:ftpID/:page": function(ftpID, page) {
			this.current = "favoriteTreatmentPlan";
			this.params = {page: page, ftpID: ftpID};
		},
		"case/visit": function() {
			this.current = "visit";
			this.params = {page: "overview"};
		},
		"case/:caseID/visit/:visitID": function(caseID, visitID) {
			this.current = "visit";
			this.params = {page: "details", caseID: caseID, visitID: visitID};
		},
	}
});

var Admin = React.createClass({displayName: "Admin",
	getDefaultProps: function() {
		var leftMenuItems = [
			{
				id: "dashboard",
				url: "",
				name: "Dashboard"
			}
		];

		if (Perms.has(Perms.DoctorsView)) {
			leftMenuItems.push({
				id: "doctorSearch",
				url: "doctors",
				name: "Doctors"
			});
		};

		if (Perms.has(Perms.DoctorsView)) {
			leftMenuItems.push({
				id: "careProviderMappings",
				url: "care_provider_mappings",
				name: "CP Mappings"
			});
		};

		if (Perms.has(Perms.ResourceGuidesView) || Perms.has(Perms.RXGuidesView)) {
			leftMenuItems.push({
				id: "guides",
				url: "guides/resources",
				name: "Guides"
			});
		}

		if (Perms.has(Perms.AnalyticsReportsView)) {
			leftMenuItems.push({
				id: "analytics",
				url: "analytics/query",
				name: "Analytics"
			});
		}

		if (Perms.has(Perms.AdminAccountsView)) {
			leftMenuItems.push({
				id: "accounts",
				url: "accounts",
				name: "Accounts"
			});
		}

		leftMenuItems.push({
			id: "drugs",
			url: "drugs",
			name: "Drugs"
		});

		if (Perms.has(Perms.PathwaysView)) {
			leftMenuItems.push({
				id: "pathways",
				url: "pathways",
				name: "Pathways"
			});
		}

		if (Perms.has(Perms.FinancialView)) {
			leftMenuItems.push({
				id: "financial",
				url: "financial/incoming",
				name: "Financial"
			})
		}

		if (Perms.has(Perms.CaseView)) {
			leftMenuItems.push({
				id: "visit",
				url: "case/visit",
				name: "Visit Overview"
			})
		}

		var rightMenuItems = [];

		return {
			leftMenuItems: leftMenuItems,
			rightMenuItems: rightMenuItems
		}
	},
	pages: {
		dashboard: function() {
			return <div className="container-fluid main"><Dashboard.Dashboard router={this.props.router} /></div>;
		},
		doctorSearch: function() {
			return <div className="container-fluid main"><Doctors.DoctorSearch router={this.props.router} /></div>;
		},
		doctor: function() {
			return <Doctors.Doctor router={this.props.router} doctorID={this.props.router.params.doctorID} page={this.props.router.params.page} />;
		},
		careProviderMappings: function() {
			return <CPMappings.CareProviderStatePathwayMappings router={this.props.router} />;
		},
		guides: function() {
			return <Guides.Guides router={this.props.router} page={this.props.router.params.page} guideID={this.props.router.params.guideID} />;
		},
		analytics: function() {
			return <Analytics.Analytics router={this.props.router} page={this.props.router.params.page} reportID={this.props.router.params.reportID} />;
		},
		accountsList: function() {
			return <Accounts.AccountList router={this.props.router} accountID={this.props.router.params.accountID} />;
		},
		account: function() {
			return <Accounts.Account router={this.props.router} accountID={this.props.router.params.accountID} page={this.props.router.params.page} />;
		},
		drugs: function() {
			return <Drugs.DrugSearch router={this.props.router} accountID={this.props.router.params.accountID} />;
		},
		pathways: function() {
			return <Pathways.Page router={this.props.router} page={this.props.router.params.page} pathwayID={this.props.router.params.pathwayID} />;
		},
		financial: function() {
			return <Financial.Page router={this.props.router} page={this.props.router.params.page} />;
		 },
		favoriteTreatmentPlan: function() {
			return <FavoriteTreatmentPlan.Page router={this.props.router} page={this.props.router.params.page} ftpID={this.props.router.params.ftpID} />;
		},
		visit: function() {
			return <Visit.Page router={this.props.router} page={this.props.router.params.page} caseID={this.props.router.params.caseID} visitID={this.props.router.params.visitID}/>;
		},
	},
	componentWillMount : function() {
		this.props.router.on("route", function() {
			this.forceUpdate();
		}.bind(this));
	},
	componentWillUnmount : function() {
		this.props.router.off("route", function() {
			this.forceUpdate();
		}.bind(this));
	},
	render: function() {
		var page = this.pages[this.props.router.current];
		return (
			<div>
				<Nav.TopNav router={this.props.router} leftItems={this.props.leftMenuItems} rightItems={this.props.rightMenuItems} activeItem={this.props.router.current} name="Admin" />
				{page ? page.bind(this)() : "Page Not Found"}
			</div>
		);
	}
});

jQuery(function() {
	var router = new AdminRouter();
	router.root = "/admin/";
	React.render(React.createElement(Admin, {
		router: router
	}), document.getElementById('base-content'));
	Backbone.history.start({
		pushState: true,
		root: router.root,
	});
});
