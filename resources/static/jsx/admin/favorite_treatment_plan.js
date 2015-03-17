var AdminAPI = require("./api.js");
var IntakeReview = require("./intake_review.js");
var Forms = require("../forms.js");
var Modals = require("../modals.js");
var Nav = require("../nav.js");
var Perms = require("./permissions.js");
var Routing = require("../routing.js");
var Utils = require("../utils.js");

module.exports = {
	Page: React.createClass({displayName: "FavoriteTreatmentPlanPage",
		menuItems: [[
			{
				id: "info",
				url: "info",
				name: "Info"
			},
			{
				id: "memberships",
				url: "memberships",
				name: "Memberships"
			}
		]],
		getDefaultProps: function() {
			return {}
		},
		info: function() {
			return <FTPInfoPage router={this.props.router} ftpID={this.props.ftpID}/>;
		},
		memberships: function() {
			return <FTPMembershipsPage router={this.props.router} ftpID={this.props.ftpID}/>;
		},
		render: function() {
			return (
				<div>
					<Nav.LeftNav router={this.props.router} items={this.menuItems} currentPage={this.props.page}>
						{this[this.props.page]()}
					</Nav.LeftNav>
				</div>
			);
		}
	})
};

var FTPInfoPage = React.createClass({displayName: "FTPInfoPage",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function() {
		return {
			ftp: null,
			ftp_fetch_error: null,
			busy: true
		};
	},
	componentWillMount: function() {
		document.title = "Favorite Treatment Plan Info";
		AdminAPI.favoriteTreatmentPlans(this.props.ftpID, function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						ftp_fetch_error: error.message,
						busy: false,
					});
					return;
				}
				this.setState({
					ftp: data.favorite_treatment_plan,
					ftp_fetch_error: null,
					busy: false,
				});
			}
		}.bind(this));
	},
	render: function() {
		return (
			<div className="container" style={{marginTop: 10}}>
				{ this.state.ftp != null ? <FTPInfoRO router={this.props.router} ftp={this.state.ftp} /> : null }
			</div>
		);
	}
});

var FTPInfoRO = React.createClass({displayName: "FTPInfoPage",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function() {
		return {};
	},
	componentWillMount: function() {},
	render: function() {
		var regimen = []
		for(var i = 0; i < this.props.ftp.regimen_plan.regimen_sections.length; ++i) {
			steps = []
			{
				for(var i2 = 0; i2 < this.props.ftp.regimen_plan.regimen_sections[i].regimen_steps.length; ++i2) {
					steps.push(<tr><td>{this.props.ftp.regimen_plan.regimen_sections[i].regimen_steps[i2].text}</td></tr>)
				}
			}
			regimen.push(
				<div className="row">
					<div className="col-md-12">
						<table className="table">
							<thead>
							<tr><th>{this.props.ftp.regimen_plan.regimen_sections[i].regimen_name}</th></tr>
							</thead>
							<tbody>
							{steps}
							</tbody>
						</table>
					</div>
				</div>)
		}

		var scheduled_messages = []
		for(var i = 0; i < this.props.ftp.scheduled_messages.length; ++i) {
			message_data = []
			message_data.push(<tr><td>Title:</td><td>{this.props.ftp.scheduled_messages[i].title}</td></tr>)
			message_data.push(<tr><td>Message:</td><td>{this.props.ftp.scheduled_messages[i].message}</td></tr>)
			message_data.push(<tr><td>Scheduled Days:</td><td>{this.props.ftp.scheduled_messages[i].scheduled_days}</td></tr>)
			for(var i2 = 0; i2 < this.props.ftp.scheduled_messages[i].attachments.length; ++i2){
				header = "Attachment " + (i2 + 1)
				message_data.push(<tr><td><strong>{header}</strong></td></tr>)
				message_data.push(<tr><td>Title:</td><td>{this.props.ftp.scheduled_messages[i].attachments[i2].title}</td></tr>)
				message_data.push(<tr><td>Type:</td><td>{this.props.ftp.scheduled_messages[i].attachments[i2].type}</td></tr>)
			}
			scheduled_messages.push(<table className="table">{message_data}</table>)
		}

		var treatments = []
		if(typeof this.props.ftp.treatment_list.treatments != "undefined"){
			for(var i = 0; i < this.props.ftp.treatment_list.treatments.length; ++i) {
				treatment_data = []
				treatment_data.push(<thead><tr><th>{this.props.ftp.treatment_list.treatments[i].drug_internal_name}</th></tr></thead>)
				treatment_data.push(<tr><td>Drug Name:</td><td>{this.props.ftp.treatment_list.treatments[i].drug_name}</td></tr>)
				treatment_data.push(<tr><td>Drug Form:</td><td>{this.props.ftp.treatment_list.treatments[i].drug_form}</td></tr>)
				treatment_data.push(<tr><td>Drug Route:</td><td>{this.props.ftp.treatment_list.treatments[i].drug_route}</td></tr>)
				treatment_data.push(<tr><td>Dispense Unit:</td><td>{this.props.ftp.treatment_list.treatments[i].dispense_unit_description}</td></tr>)
				treatment_data.push(<tr><td>Dispense Value:</td><td>{this.props.ftp.treatment_list.treatments[i].dispense_value}</td></tr>)
				treatment_data.push(<tr><td>Dosage Strength:</td><td>{this.props.ftp.treatment_list.treatments[i].dosage_strength}</td></tr>)
				treatment_data.push(<tr><td>Refills:</td><td>{this.props.ftp.treatment_list.treatments[i].refills}</td></tr>)
				treatment_data.push(<tr><td>Patient Instructions:</td><td>{this.props.ftp.treatment_list.treatments[i].patient_instructions}</td></tr>)
				treatment_data.push(<tr><td>Substitutions Allowed:</td><td>{this.props.ftp.treatment_list.treatments[i].substitutions_allowed ? "true" : "false"}</td></tr>)
				treatments.push(<table className="table">{treatment_data}</table>)
			}
		}

		var resource_guides = []
		if(typeof this.props.ftp.resource_guides != "undefined"){
			for(var i = 0; i < this.props.ftp.resource_guides.length; ++i) {
				resource_guides.push(<div className="col-md-12"><FTPResourceGuide router={this.props.router} guide={this.props.ftp.resource_guides[i]} /></div>)
			}
		}
		return (
			<div className="container" style={{marginTop: 10}}>
				<div className="row">
					<div className="col-sm-12 col-md-12 col-lg-9">
						<h1>{this.props.ftp.name}</h1>
					</div>
				</div>
				<div className="row">
					<div className="col-md-12">
						<h2>Note:</h2>
						<pre>{this.props.ftp.note}</pre>
					</div>
				</div>
				<div className="row">
					<div className="col-md-12">
						<h2>Treatments:</h2>
						<table className="table">
							{treatments}
						</table>
					</div>
				</div>
				<div className="row">
					<div className="col-md-12">
						<h2>Regimen:</h2>						
						{regimen}
					</div>
				</div>
				<div className="row">
					<div className="col-md-12">
						<h2>Resource Guides:</h2>						
						{resource_guides}
					</div>
				</div>
				<div className="row">
					<div className="col-md-12">
						<h2>Scheduled Messages:</h2>
						<table className="table">
							{scheduled_messages}
						</table>
					</div>
				</div>
			</div>
		);
	}
});

var FTPResourceGuide = React.createClass({displayName: "FTPResourceGuide",
	mixins: [Routing.RouterNavigateMixin],
	render: function() {
		return (
			<div key={"guide-"+this.props.guide.id} className="item">
				<img src={this.props.guide.photo_url} width="32" height="32" />
				&nbsp;<a href={"/admin/guides/resources/"+this.props.guide.id} onClick={this.onNavigate}>{this.props.guide.title || "NO TITLE"}</a>
			</div>
		);
	}
});

var FTPMembershipsPage = React.createClass({displayName: "FTPMembershipsPage",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function() {
		return {
			membershipData: null,
			membershipFetchError: null,
			busy: true
		};
	},
	componentWillMount: function() {
		document.title = "Favorite Treatment Plan Memberships";
		this.queryMemberships()
	},
	updatedMemberships: function() {
		this.queryMemberships()
	},
	queryMemberships: function(){
		AdminAPI.favoriteTreatmentPlanMemberships(this.props.ftpID, function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						membershipFetchError: error.message,
						busy: false,
					});
					return;
				}
				data.memberships.sort(function(a,b){
					if(a.first_name+a.last_name < b.first_name+b.last_name) return -1
					if(a.first_name+a.last_name > b.first_name+b.last_name) return 1
					return 0
				})
				this.setState({
					membershipData: data,
					membershipFetchError: null,
					busy: false,
				});
			}
		}.bind(this));
	},
	render: function() {
		return (
			<div className="container" style={{marginTop: 10}}>
				{(Perms.has(Perms.FTPEdit) && this.state.membershipData != null) ? <FTPAddMembershipModal memberships={this.state.membershipData.memberships} ftpID={this.props.ftpID} onUpdate={this.updatedMemberships} /> : null}
				<div className="row">
					{
						this.state.membershipData != null ?
						<div className="row">
							<div className="col-sm-12 col-md-12 col-lg-9">
								{Perms.has(Perms.FTPEdit) ? <div className="pull-right"><button className="btn btn-default" data-toggle="modal" data-target="#add-membership-modal">+</button></div> : null}
								<h1>{this.state.membershipData.name}</h1>
								<FTPMembershipList router={this.props.router} memberships={this.state.membershipData.memberships} onUpdate={this.updatedMemberships}/>
							</div>
						</div> : null
					}
				</div>
			</div>
		);
	}
});

var FTPMembershipList = React.createClass({displayName: "FTPMembershipList",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function() {
		return {
			memberships: this.props.memberships,
			membership_fetch_error: null,
		};
	},
	componentWillReceiveProps: function(nextProps) {
  	this.setState({
    	memberships: nextProps.memberships
		})
	},
	onRemove: function(membership, e) {
		e.preventDefault()
		AdminAPI.deleteFavoriteTreatmentPlanMembership(membership.favorite_treatment_plan_id, membership.doctor_id, membership.pathway_tag, function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						membership_delete_error: error.message,
					});
					return;
				}
				this.props.onUpdate()
			}
		}.bind(this));
	},
	closeOnAdd: function(e) {
		e.preventDefault()
		this.setState({
			adding: false
		});
	},
	render: function() {
		var memberships = []
		for(var i = 0; i < this.state.memberships.length; ++i){
			memberships.push(
				<tr>
				<td>
				<a href={"/admin/doctors/" + this.state.memberships[i].doctor_id + "/favorite_treatment_plans"} onClick={this.onNavigate}>
					{this.state.memberships[i].first_name + " " + this.state.memberships[i].last_name}
				</a>
				</td>
				<td>{this.state.memberships[i].pathway_name}</td>
				{Perms.has(Perms.FTPEdit) ? <td><a href="#" onClick={this.onRemove.bind(this, this.state.memberships[i])}><span className="glyphicon glyphicon-remove" /></a></td> : null}
				</tr>
			)
		}

		return (
			<div>
			<table className="table"><tbody>{memberships}</tbody></table>
			</div>
		);
	}
});

var FTPAddMembershipModal = React.createClass({displayName: "FTPAddMembershipModal",
	getInitialState: function(){
		return {
			doctors: [],
			doctor_fetch_error: null,
			query: "",
			queuedMemberships: {}
		}
	},
	componentWillMount: function() {
		AdminAPI.pathways(true, function(success, data, error) {
			if (!success) {
				this.setState({busy: false, doctor_fetch_error: error.message});
				return;
			}
			pathwaySelectOptions = []
			pathwayTagByID = {}
			initalTag = data.pathways[0].tag
			for(var i = 0; i < data.pathways.length; ++i) {
				pathwaySelectOptions.push(<option value={data.pathways[i].tag}>{data.pathways[i].name}</option>)
				pathwayTagByID[data.pathways[i].id] = data.pathways[i].tag
			}
			existingMembershipsByTag = {}
			for(var i = 0; i < this.props.memberships.length; ++i) {
				tag = pathwayTagByID[this.props.memberships[i].pathway_id]
				if(typeof existingMembershipsByTag[tag] == "undefined") {
					existingMembershipsByTag[tag] = {}
				}
				existingMembershipsByTag[tag][this.props.memberships[i].doctor_id] = true
			}
			this.setState({
				pathwaySelectOptions: pathwaySelectOptions,
				existingMembershipsByTag: existingMembershipsByTag,
				initialPathwayTag: initalTag,
			})
		}.bind(this));
	},
	onQueryChange: function(e) {
		this.setState({query: e.target.value})
		this.search(e.target.value)
	},
	onSearchSubmit: function(e) {
		e.preventDefault()
		this.search(this.state.query)
	},
	search: function(q) {
		if(q != ""){
			this.setState({busy: true, error: null});
			AdminAPI.searchDoctors(q, function(success, res, error) {
				if (this.isMounted()) {
					if (!success) {
						this.setState({busy: false, doctor_fetch_error: error.message});
						return;
					}
					this.setState({busy: false, doctor_fetch_error: null, doctors: res.results || []});
				}
			}.bind(this));
		} else {
			this.setState({busy: false, doctor_fetch_error: null, doctors: []})
		}
	},
	onAddSave: function(){
		var requests = []
		var queuedMemberships = this.state.queuedMemberships
		for(var doctorID in queuedMemberships){
			for(var pathwayTag in queuedMemberships[doctorID]){
				requests.push({doctor_id: doctorID, pathway_tag: pathwayTag})
			}
		}
		if(requests.length > 0){
			AdminAPI.createFavoriteTreatmentPlanMemberships(this.props.ftpID, requests, function(success, res, error) {
				if (this.isMounted()) {
					if (!success) {
						this.setState({busy: false, membershipSaveError: error.message});
						return;
					}
					this.setState({busy: false, membershipSaveError: null});
					this.props.onUpdate()
					$("#modal-ftp-membership").modal('hide');
				}
			}.bind(this));
		} else {
			$("#modal-ftp-membership").modal('hide');
		}
	},
	onAddMembership: function(doctorID, pathwayTag){
		var queuedMemberships = this.state.queuedMemberships
		if(queuedMemberships[doctorID] == null) {
			queuedMemberships[doctorID] = {}
		}
		queuedMemberships[doctorID][pathwayTag] = true
		this.setState({
			queuedMemberships: queuedMemberships,
		})
	},
	onRemoveMembership: function(doctorID, pathwayTag){
		var queuedMemberships = this.state.queuedMemberships
		delete(queuedMemberships[doctorID][pathwayTag])
		this.setState({
			queuedMemberships: queuedMemberships,
		})
	},
	render: function() {
		var doctors = []
		this.state.doctors.map(function(){
			doctors.push(
				<FTPAddableMembership 
					doctor={this.state.doctors[i]} 
					pathwaySelectOptions={this.state.pathwaySelectOptions} 
					existingMembershipsByTag={existingMembershipsByTag} 
					initialPathwayTag={this.state.initialPathwayTag} 
					onAddMembership={this.onAddMembership} 
					onRemoveMembership={this.onRemoveMembership}/>
			)
		})
		return (
			<Modals.ModalForm contentClassName="modal-ftp-membership" id="add-membership-modal" title="Add Membership"
				cancelButtonTitle="Cancel" submitButtonTitle="Save"
				onSubmit={this.onAddSave}>
				<div>
					<div className="form-group">
						<input autofocus type="text" className="form-control" name="q" value={this.state.query} onChange={this.onQueryChange} />
						{
							<table className="table">
								<tbody>{doctors}</tbody>
							</table>
						}
					</div>
				</div>
			</Modals.ModalForm>
		)
	}
});

var FTPAddableMembership = React.createClass({displayName: "FTPAddableMembership",
	getInitialState: function(){
		return {
			added: false,
			canBeAdded: false,
			selectedPathway: this.props.initialPathwayTag,
			selectHistory: {}
		}
	},
	componentWillMount: function() {
		this.setState({
			canBeAdded: this.props.existingMembershipsByTag[this.props.initialPathwayTag] == null || this.props.existingMembershipsByTag[this.props.initialPathwayTag][this.props.doctor.doctor_id] != true
		})
	},
	onSelectChange: function(e){
		e.preventDefault()
		this.setState({
			selectedPathway: e.target.value,
			canBeAdded: this.props.existingMembershipsByTag[e.target.value] == null || this.props.existingMembershipsByTag[e.target.value][this.props.doctor.doctor_id] != true,
			added: selectHistory[this.props.doctor.doctor_id][e.target.value] == true
		})
	},
	onAdd: function(e){
		e.preventDefault()
		if(this.state.canBeAdded){
			selectHistory = this.state.selectHistory
			if(selectHistory[this.props.doctor.doctor_id] == null) {
				selectHistory[this.props.doctor.doctor_id] = {}
			}
			selectHistory[this.props.doctor.doctor_id][this.state.selectedPathway] = true
			this.props.onAddMembership(this.props.doctor.doctor_id, this.state.selectedPathway)
			this.setState({
				added: true,
				selectHistory: selectHistory,
			})
		}
	},
	onRemove: function(e){
		e.preventDefault()
		if(this.state.canBeAdded){
			selectHistory = this.state.selectHistory
			delete(selectHistory[this.props.doctor.doctor_id][this.state.selectedPathway])
			this.props.onRemoveMembership(this.props.doctor.doctor_id, this.state.selectedPathway)
			this.setState({
				added: false,
				selectHistory: selectHistory,
			})
		}
	},
	render: function(){
		return(
			<tr>
			<td>
				<a href={"/admin/doctors/" + this.props.doctor.doctor_id + "/favorite_treatment_plans"} onClick={this.onNavigate}>
					{this.props.doctor.first_name + " " + this.props.doctor.last_name}</a>
			</td>
			<td>{this.props.doctor.email}</td>
			<td><select onChange={this.onSelectChange}>{this.props.pathwaySelectOptions}</select></td>
			<td>
				<a href="#">
					{ 
						this.state.canBeAdded ? 
							this.state.added ? 
								<span title="This doctor has been selected to be added to this FTP/Pathway combination" className="glyphicon glyphicon-check" onClick={this.onRemove} /> : 
								<span title="Select this doctor to be added to this FTP/Pathway combination" className="glyphicon glyphicon-plus" onClick={this.onAdd} /> : 
									<span title="This doctor already has access to this FTP/Pathway combination" className="glyphicon glyphicon-lock" />
					}
				</a>
			</td>
			</tr>
		)
	}
});