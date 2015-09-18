/* @flow */

var AdminAPI = require("./api.js");
var Forms = require("../../libs/forms.js");
var Modals = require("../../libs/modals.js");
var Nav = require("../../libs/nav.js");
var React = require("react");
var Routing = require("../../libs/routing.js");
var Utils = require("../../libs/utils.js");
var Perms = require("./permissions.js");
require('date-utils');

module.exports = {
	Page: React.createClass({displayName: "VisitPage",
		menuItems: [[
			{
				id: "visit",
				url: "/admin/case/visit",
				name: "Untreated Visits"
			}
		]],
		pages: {
			overview: function(): any {
				return <VisitOverviewPage router={this.props.router} />;
			},
			details: function(): any {
				return <VisitDetailsPage router={this.props.router} caseID={this.props.caseID} visitID={this.props.visitID} />;
			},
		},
		getDefaultProps: function(): any {
			return {}
		},
		render: function(): any {
			return (
				<div>
					<Nav.LeftNav router={this.props.router} items={this.menuItems} currentPage={this.props.page}>
						{this.pages[this.props.page].bind(this)()}
					</Nav.LeftNav>
				</div>
			);
		}
	})
};

var VisitOverviewPage = React.createClass({displayName: "VisitOverviewPage",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {showAssigned: false};
	},
	componentWillMount: function() {
		document.title = "Visit Overview";
	},
	toggleShowAssigned: function() {
		this.setState({
			showAssigned: !this.state.showAssigned,
		})
	},
	render: function(): any {
		return (
			<div className="container" style={{marginTop: 10}}>
				{Perms.has(Perms.CaseView) ? <a href="#" onClick={this.toggleShowAssigned}>{this.state.showAssigned ? "[Hide Assigned]" : "[Show Assigned]"}</a> : null }
				{Perms.has(Perms.CaseView) ? <VisitSummaryList router={this.props.router} visitStatus="uncompleted" showAssigned={this.state.showAssigned} /> : null }
			</div>
		);
	}
});

var VisitSummaryList = React.createClass({displayName: "VisitSummaryList",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {
			busy: true,
			summaries: []
		};
	},
	componentWillMount: function() {
		AdminAPI.visitSummaries(this.props.visitStatus, function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						visitOverviewError: error.message,
						busy: false,
					});
					return;
				}
				this.setState({
					summaries: data.visit_summaries || [],
					visitOverviewError: null,
					busy: false,
				});
			}
		}.bind(this));
	},
	render: function(): any {
		return (
			<div className="container" style={{marginTop: 10}}>
			{this.state.busy ? <Utils.LoadingAnimation /> : null}
			{this.state.visitOverviewError ? <Utils.Alert type="danger">{this.state.visitOverviewError}</Utils.Alert> : null}
			{this.state.summaries.map(function(v){return (this.props.showAssigned || v.doctor_id == null) ? <VisitSummaryCard key={v.visit_id} router={this.props.router} summary={v} /> : null}.bind(this))}
			</div>
		);
	}
});

var LockTypeDisplayStatus = {"TEMP": "Temporary", "ACTIVE": "Permanent"}

var VisitSummaryCard = React.createClass({ displayName: "VisitSummaryCard",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {
			busy: true,
		};
	},
	componentWillMount: function() {
		this.setState({
			busy: false,
		});
	},
	render: function(): any {
		var visitSubmitted = new Date(0)
		var lockTaken = new Date(0)
		var currentEpoch = new Date().getTime() / 1000
		visitSubmitted.setUTCSeconds(this.props.summary.submitted_epoch)
		lockTaken.setUTCSeconds(this.props.summary.lock_taken_epoch)
		return (
			<div className="card">
			<table className="table">
				<tbody>
					<tr>
						<td><div className="initials-circle"><a href={"/case/" + this.props.summary.case_id + "/visit/" + this.props.summary.visit_id} onClick={this.onNavigate}>{this.props.summary.patient_initials}</a></div></td><td><div className="card-title">{this.props.summary.case_name}</div></td><td><div className="card-title">{this.props.summary.submission_state}</div></td>
					</tr>
					<tr>
						<td>Visit Submitted:</td>
						{this.props.summary.submitted_epoch == 0 ?
							<td><div className="alert-text">Unsubmitted</div></td> :
							<td>{visitSubmitted.toString()}</td>}
						{this.props.summary.submitted_epoch == 0 ?
							null :
							<td>{Utils.timeSince(visitSubmitted.getTime()/1000, currentEpoch)}</td>}
					</tr>
					<tr>
						<td>Doctor:</td><td>{this.props.summary.doctor_id != null ? <div className="success-text">{this.props.summary.doctor_with_lock}</div> : <div className="alert-text">Unassigned</div>}</td><td><strong>{this.props.summary.first_available ? "First Available" : "Doctor Selected"}</strong></td>
					</tr>
					{
						this.props.summary.doctor_id != null ?
						<tr>
							<td>Lock Acquired:</td><td>{lockTaken.toString()}</td><td>{Utils.timeSince(lockTaken.getTime()/1000, currentEpoch)}</td>
						</tr> : null
					}
					{
						this.props.summary.doctor_id != null ?
						<tr>
							<td>Lock Type:</td><td>{this.props.summary.lock_type == "TEMP" ? <div className="alert-text">{LockTypeDisplayStatus[this.props.summary.lock_type]}</div> : LockTypeDisplayStatus[this.props.summary.lock_type]}</td>
						</tr> : null
					}
					<tr>
						<td>Visit Type:</td><td>{this.props.summary.type}</td>
					</tr>
					<tr>
						<td>Internal Visit Status:</td><td>{this.props.summary.status}</td>
					</tr>
				</tbody>
			</table>
			</div>
		);
	}
});

var VisitDetailsPage = React.createClass({displayName: "VisitDetailsPage",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {};
	},
	componentWillMount: function() {
		document.title = "Visit Details";
	},
	render: function(): any {
		return (
			<div className="container" style={{marginTop: 10}}>
				<h1>Visit Details</h1>
				<VisitSummary router={this.props.router} caseID={this.props.caseID} visitID={this.props.visitID}/>
			</div>
		);
	}
});

var VisitSummary = React.createClass({displayName: "VisitSummary",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {busy: true};
	},
	componentWillMount: function() {
		AdminAPI.visitSummary(this.props.caseID, this.props.visitID, function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						visitDetailsError: error.message,
						busy: false,
					});
					return;
				}
				this.setState({
					summary: data.visit_summary,
					visitDetailsError: null,
					busy: false,
				});
			}
		}.bind(this));
	},
	render: function(): any {
		if(!this.state.busy) {
			var visitSubmitted = new Date(0)
			var lockTaken = new Date(0)
			var currentEpoch = new Date().getTime() / 1000
			visitSubmitted.setUTCSeconds(this.state.summary.submitted_epoch)
			lockTaken.setUTCSeconds(this.state.summary.lock_taken_epoch)
		}
		return (
			<div className="container" style={{marginTop: 10}}>
				{
					this.state.busy ?
						<img src={Utils.staticURL("/img/loading.gif")} /> :
						this.state.visitDetailsError == null ?
							<table className="table">
								<tbody>
									<tr>
										<td><div className="initials-circle">{this.state.summary.patient_initials}</div></td><td><div className="card-title">{this.state.summary.case_name}</div></td><td><div className="card-title">{this.state.summary.submission_state}</div></td>
									</tr>
									<tr>
										<td>Visit Submitted:</td>
										{this.state.summary.submitted_epoch == 0 ?
											<td><div className="alert-text">Unsubmitted</div></td> :
											<td>{visitSubmitted.toString()}</td>}
										{this.state.summary.submitted_epoch == 0 ?
											null :
											<td>{Utils.timeSince(visitSubmitted.getTime()/1000, currentEpoch)}</td>}
									</tr>
									<tr>
										<td>Doctor:</td><td>{this.state.summary.doctor_id != null ? <div className="success-text">{this.state.summary.doctor_with_lock}</div> : <div className="alert-text">Unassigned</div>}</td><td><strong>{this.state.summary.first_available ? "First Available" : "Doctor Selected"}</strong></td>
									</tr>
									{
										this.state.summary.doctor_id != null ?
										<tr>
											<td>Lock Acquired:</td><td>{lockTaken.toString()}</td><td>{Utils.timeSince(lockTaken.getTime()/1000, currentEpoch)}</td>
										</tr> : null
									}
									{
										this.state.summary.doctor_id != null ?
										<tr>
											<td>Lock Type:</td><td>{this.state.summary.lock_type == "TEMP" ? <div className="alert-text">{LockTypeDisplayStatus[this.state.summary.lock_type]}</div> : LockTypeDisplayStatus[this.state.summary.lock_type]}</td>
										</tr> : null
									}
									<tr>
										<td>Visit Type:</td><td>{this.state.summary.type}</td>
									</tr>
									<tr>
										<td>Internal Visit Status:</td><td>{this.state.summary.status}</td>
									</tr>
									<tr>
										<td>Visit ID:</td><td>{this.state.summary.visit_id}</td>
									</tr>
									<tr>
										<td>Case ID:</td><td>{this.state.summary.case_id}</td>
									</tr>
								</tbody>
							</table> :
							<Utils.Alert type="danger">{this.state.visitDetailsError}</Utils.Alert>
				}
			</div>
		);
	}
});