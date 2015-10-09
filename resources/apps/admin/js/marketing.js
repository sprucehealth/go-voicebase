/* @flow */

var AdminAPI = require("./api.js");
var Forms = require("../../libs/forms.js");
var Modals = require("../../libs/modals.js");
var Nav = require("../../libs/nav.js");
var React = require("react");
var Routing = require("../../libs/routing.js");
var Utils = require("../../libs/utils.js");
var Time = require("../../libs/time.js");
var Perms = require("./permissions.js");
require('date-utils');

module.exports = {
	Page: React.createClass({displayName: "MarketingPage",
		menuItems: [[
			{
				id: "promotions",
				url: "/marketing/promotions",
				name: "Promotions"
			},
			{
				id: "referral_templates",
				url: "/marketing/promotions/referral_templates",
				name: "Promotion Referral Template"
			},
			{

				id: "referral_routes",
				url: "/marketing/promotions/referral_routes",
				name: "Promotion Referral Routes"
			},
			{
				id: "in_app_feedback",
				url: "/marketing/promotions/feedback",
				name: "In-app Feedback"
			}
		]],
		pages: {
			promotions: function(): any {
				return <PromotionOverview router={this.props.router} />;
			},
			referral_routes: function(): any {
				return <PromotionReferralRouteOverview router={this.props.router} />;
			},
			referral_templates: function(): any {
				return <PromotionReferralTemplateOverview router={this.props.router} />;
			},
			feedback: function(): any {
				return <InAppFeedbackPage router={this.props.router} />;
			},
			template: function(): any {
				var updating = (typeof(this.props.templateID) !== "undefined");
				return <CreateFeedbackTemplatePage router={this.props.router} updating={updating} templateID={this.props.templateID} />;
			}
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

// BEGIN: Promotion Management
var AddPromotionModal = React.createClass({displayName: "AddPromotionModal",
	// Eventually these should be fetched from an API,
	promoTypes: [{name: "Percent Off", value:"promo_percent_off"}, {name: "Money Off", value: "promo_money_off"}],
	valueDescriptorByType: {"promo_money_off": "cents", "promo_percent_off": "%"},
	getInitialState: function(): any {
		var groups = [{name: "", value: ""}]
		return {
			error: null,
			busy: false,
			promoTypes: this.promoTypes,
			promoTypesSelectedValue: this.promoTypes[0].value,
			groups: groups,
			groupSelectedValue: groups[0].value,
		};
	},
	componentWillMount: function(): any {
		document.title = "Marketing | Promotions";
		AdminAPI.promotionGroups(function(success, data, error){
			if(success) {
				var groups = []
				for(var i = 0; i < data.promotion_groups.length; i++){
					groups.push({name: data.promotion_groups[i].name + " - Max Allowed " + data.promotion_groups[i].max_allowed_promos.toString(), value: data.promotion_groups[i].name})
				}
				if(groups.length > 0) {
					this.setState({groups: groups, groupSelectedValue: groups[0].value})
				}
			} else {
				this.setState({error: error.message})
			}
		}.bind(this))
	},
	onPromoTypeChange: function(e): any {
		e.preventDefault();
		var state = {
			promoTypesSelectedValue: e.target.value,
			value: this.state.val,
		}
		if(!this.valueDescriptorByType[e.target.value]) {
			state.value = null;
		}
		this.setState(state);
	},
	onGroupChange: function(e): any {
		e.preventDefault();
		this.setState({
			groupSelectedValue: e.target.value,
		});
	},
	onImageURLChange: function(e): any {
		e.preventDefault();
		this.setState({
			imageURL: e.target.value,
		});
	},
	onImageWidthChange: function(e): any {
		e.preventDefault();
		var state = {
			imageWidth: e.target.value,
			error: null,
		};
		if(!Utils.isInteger(e.target.value) && e.target.value != "") {
			state.error = "Image Width must be an Integer"; 
		}
		this.setState(state);
	},
	onImageHeightChange: function(e): any {
		e.preventDefault();
		var state = {
			imageHeight: e.target.value,
			error: null,
		};
		if(!Utils.isInteger(e.target.value) && e.target.value != "") {
			state.error = "Image Height must be an Integer"; 
		}
		this.setState(state);
	},
	onPromoCodeChange: function(e): any {
		e.preventDefault();
		this.setState({
			promoCode: e.target.value,
		});
	},
	onDisplayMessageChange: function(e): any {
		e.preventDefault();
		this.setState({
			displayMessage: e.target.value,
		});
	},
	onLineItemDescriptionChange: function(e): any {
		e.preventDefault();
		this.setState({
			lineItemDescription: e.target.value,
		});
	},
	onSuccessMessageChange: function(e): any {
		e.preventDefault();
		this.setState({
			successMessage: e.target.value,
		});
	},
	onValueChange: function(e): any {
		e.preventDefault();
		var state = {
			val: e.target.value,
			error: null,
		};
		if(!Utils.isInteger(e.target.value) && e.target.value != "") {
			state.error = "Value must be an Integer"; 
		}
		this.setState(state);
	},
	onExpiresChange: function(e): any {
		e.preventDefault();
		this.setState({
			expires: e.target.value,
		});
	},
	onAdd: function(): any {
		if(this.validateSubmitState()) {
			var expires = null
			if(this.state.expires) {
				expires = new Date(this.state.expires)
			}
			AdminAPI.addPromotion(
				this.state.promoCode, 
				this.state.promoTypesSelectedValue, 
				this.state.groupSelectedValue, 
				{
					display_msg: this.state.displayMessage, 
					short_msg: this.state.lineItemDescription,
					success_msg: this.state.successMessage,
					image_url: this.state.imageURL,
					image_width: parseInt(this.state.imageWidth),
					image_height: parseInt(this.state.imageHeight),
					value: parseInt(this.state.val),
					group: this.state.groupSelectedValue,
					type: this.state.promoTypesSelectedValue,
				}, 
				expires, 
				function(success, data, error){
					if (this.isMounted()) {
						if (!success) {
							this.setState({
								error: error.message
							});
						} else {
							this.props.onSuccess();
							$("#add-promotion-modal").modal('hide');
						}
					}
				}.bind(this));
		}
		return true;
	},
	validateSubmitState: function(): boolean {
		if(!this.state.promoTypesSelectedValue) {
			this.setState({error: "promoType required"});
		} else if (!this.state.groupSelectedValue) {
			this.setState({error: "group required"});
		} else if (!this.state.promoCode) {
			this.setState({error: "promoCode required"});
		} else if (this.state.imageURL && (!this.state.imageWidth || !this.state.imageHeight)) {
			this.setState({error: "imageWidth and imageHeight required if imageURL present"});
		} else if (this.state.imageWidth && !Utils.isInteger(this.state.imageWidth)) {
			this.setState({error: "imageWidth must be an Integer"});
		} else if (this.state.imageHeight && !Utils.isInteger(this.state.imageHeight)) {
			this.setState({error: "imageHeight must be an Integer"});
		} else if (!this.state.displayMessage) {
			this.setState({error: "displayMessage required"});
		} else if (!this.state.lineItemDescription) {
			this.setState({error: "lineItemDescription required"});
		} else if (!this.state.successMessage) {
			this.setState({error: "successMessage required"});
		} else if (this.valueDescriptorByType[this.state.promoTypesSelectedValue] && !this.state.val) {
			this.setState({error: "value required"});
		} else if (this.valueDescriptorByType[this.state.promoTypesSelectedValue] && !Utils.isInteger(this.state.val)) {
			this.setState({error: "value must be an Integer"});
		} else {
			return true;
		}
		return false;
	},
	render: function(): any {
		return (
			<Modals.ModalForm id="add-promotion-modal" title="Add Promotion"
				cancelButtonTitle="Cancel" submitButtonTitle="Add"
				onSubmit={this.onAdd}>

				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}

				<Forms.FormSelect label="Promo Type" onChange={this.onPromoTypeChange} value={this.state.promoTypesSelectedValue} opts={this.state.promoTypes} />
				<Forms.FormSelect label="Eligible Group" onChange={this.onGroupChange} value={this.state.groupSelectedValue} opts={this.state.groups} />
				<Forms.FormInput label="Promo Code" value={this.state.promoCode} onChange={this.onPromoCodeChange} />
				<Forms.FormInput label="*Image URL" value={this.state.imageURL} onChange={this.onImageURLChange} />
				{this.state.imageURL ? <img src={this.state.imageURL} /> : null}
				{this.state.imageURL ? <Forms.FormInput label="Image Width" value={this.state.imageWidth} onChange={this.onImageWidthChange} /> : null}
				{this.state.imageURL ? <Forms.FormInput label="Image Height" value={this.state.imageHeight} onChange={this.onImageHeightChange} /> : null}
				<Forms.FormInput label="Display Message" value={this.state.displayMessage} onChange={this.onDisplayMessageChange} />
				<Forms.FormInput label="Line Item Description" value={this.state.lineItemDescription} onChange={this.onLineItemDescriptionChange} />
				<Forms.FormInput label="Success Message" value={this.state.successMessage} onChange={this.onSuccessMessageChange} />
				{this.valueDescriptorByType[this.state.promoTypesSelectedValue] ? <Forms.FormInput label={"Value - (" + this.valueDescriptorByType[this.state.promoTypesSelectedValue] + ")"} value={this.state.val} onChange={this.onValueChange} /> : null}
				<Forms.FormDatePicker label="*Expires" value={this.state.expires} onChange={this.onExpiresChange} />
			</Modals.ModalForm>
		);
	}
});

var UpdatePromotionModal = React.createClass({displayName: "UpdatePromotionModal",
	getInitialState: function(): any {
		return {
			error: null,
			busy: false,
			expires: null,
		};
	},
	onSubmit: function(): any {
		this.setState({
			busy: true
		});
		AdminAPI.updatePromotion(this.props.promotion.code_id, this.state.expires && this.state.expires != "" ? new Date(this.state.expires).getTime() / 1000 : null, function(success, data, error){
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message,
						busy: false,
					});
				} else {
					this.setState({
						error: null,
						busy: false,
					});
					this.props.onSuccess();
					$("#update-promotion-modal").modal('hide');
				}
			}
		}.bind(this));
		return true;
	},
	onExpiresChange: function(e): any {
		e.preventDefault();
		this.setState({
			expires: e.target.value,
		});
	},
	render: function(): any {
		return (
			<Modals.ModalForm id="update-promotion-modal" title={this.props.promotion ? "Update Promotion " + this.props.promotion.code + "?" : ""}
				cancelButtonTitle="Cancel" submitButtonTitle="Update"
				onSubmit={this.onSubmit}>

				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				<Forms.FormDatePicker label="*Expires" value={this.state.expires} onChange={this.onExpiresChange} />
			</Modals.ModalForm>
		);
	}
});

var PromotionOverview = React.createClass({displayName: "PromotionOverview",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {
			types: [],
			showExpired: false,
		};
	},
	componentWillMount: function() {
		document.title = "Marketing | Promotions";
		this.fetchPromotions();
	},
	fetchPromotions: function() {
		AdminAPI.promotions(this.state.types, function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message,
						busy: false,
					});
					return;
				}
				this.setState({
					promotions: data.promotions,
					error: null,
					busy: false,
				});
			}
		}.bind(this));
	},
	onFilterChange: function(types: Array<string>){
		this.setState({
			types: types,
		})
	},
	toggleExpired: function(e) {
		e.preventDefault();
		this.setState({
			showExpired: !this.state.showExpired
		})
	},
	setPromotionForUpdate: function(promotion) {
		this.setState({
			promotionForUpdate: promotion
		})
	},
	render: function(): any {
		return (
			<div className="container-flow" style={{marginTop: 10}}>
				{Perms.has(Perms.MarketingEdit) ? <AddPromotionModal onSuccess={this.fetchPromotions} /> : null}
				{Perms.has(Perms.MarketingEdit) ? <UpdatePromotionModal onSuccess={this.fetchPromotions} promotion={this.state.promotionForUpdate}/> : null}
				<h2>Promotions</h2>
				{Perms.has(Perms.MarketingView) ? <a href="#" onClick={this.toggleExpired}>{this.state.showExpired ? "[Hide Expired]" : "[Show Expired]"}</a>: null }
				{Perms.has(Perms.MarketingEdit) ? <div className="pull-right"><button className="btn btn-default" data-toggle="modal" data-target="#add-promotion-modal">+</button></div> : null}
				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				{(Perms.has(Perms.MarketingView) &&	this.state.promotions)? <PromotionList router={this.props.router} promotions={this.state.promotions} showExpired={this.state.showExpired} setForUpdate={this.setPromotionForUpdate}/> : null}
			</div>
		);
	}
});

var PromotionList = React.createClass({displayName: "PromotionList",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {
			promotions: null,
		};
	},
	render: function(): any {
		return (
			<div>
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				{this.props.promotions ? this.props.promotions.map(function(p){
					return (this.props.showExpired || p.expires == null || !Time.isTimestampBeforeNow(p.expires)) ? <Promotion key={p.code_id} promotion={p} setForUpdate={this.props.setForUpdate}/> : null
					}.bind(this)) : null
				}
			</div>
		);
	}
});

var Promotion = React.createClass({displayName: "Promotion",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {};
	},
	onEdit: function(promotion, e): any {
		e.preventDefault();
		this.props.setForUpdate(promotion)
		$("#update-promotion-modal").modal('show');
	},
	render: function(): any {
		return (
			<div className="card">
				<table className="table">
					<tbody>
						<tr><td>Promotion Code ID:</td><td>{this.props.promotion.code_id}</td><td><a href="#" onClick={this.onEdit.bind(this,this.props.promotion)}>Edit</a></td></tr>
						<tr><td>Group:</td><td>{this.props.promotion.data.group}</td><td></td></tr>
						<tr><td>Message:</td><td>{this.props.promotion.data.display_msg}</td><td></td></tr>
						<tr><td>Line Item Description:</td><td>{this.props.promotion.data.short_msg}</td><td></td></tr>
						<tr><td>Success Message:</td><td>{this.props.promotion.data.success_msg}</td><td></td></tr>
						{this.props.promotion.data.image_url ? <tr><td>Image:</td><td><img src={this.props.promotion.data.image_url} /></td><td></td></tr> : null}
						{this.props.promotion.data.image_url ? <tr><td>Image Height:</td><td>{this.props.promotion.data.image_width || "Undefined"} </td><td></td></tr> : null}
						{this.props.promotion.data.image_url ? <tr><td>Image Width:</td><td>{this.props.promotion.data.image_height || "Undefined"} </td><td></td></tr> : null}
						<tr><td>Code:</td><td>{this.props.promotion.code}</td><td></td></tr>
						{this.props.promotion.data.value != 0 ? <tr><td>Type:</td><td>{this.props.promotion.type}</td><td></td></tr> : null}
						{this.props.promotion.data.value != 0 ? <tr><td>Type:</td><td>{this.props.promotion.data.value}</td><td></td></tr> : null}
						<tr><td>Created:</td><td>{(new Date(this.props.promotion.created * 1000)).toString()}</td><td></td></tr>
						<tr><td>Expires:</td><td>{this.props.promotion.expires ? (new Date(this.props.promotion.expires * 1000)).toString() : "Never"}</td><td></td></tr>
					</tbody>
				</table>
			</div>
		);
	}
})
// END: Promotion Management

// BEGIN: Promotion Referral Route Management
var AddPromotionReferralRouteModal = React.createClass({displayName: "AddPromotionReferralRouteModal",
	// Eventually this should be fetched from the server
	genders: [{name: "", value: "none"}, {name: "Male", value:"M"}, {name: "Female", value: "F"}],
	getInitialState: function(): any {
		return {
			error: null,
			busy: false,
			gendersSelectedValue: null,
		};
	},
	onPriorityChange: function(e): any {
		e.preventDefault();
		var state = {
			priority: e.target.value,
			error: null,
		};
		if(!Utils.isInteger(e.target.value) && e.target.value != "") {
			state.error = "Priority must be an Integer"; 
		}
		this.setState(state);
	},
	onPromotionCodeIDChange: function(e): any {
		e.preventDefault();
		var state = {
			promotionCodeID: e.target.value,
			error: null,
		};
		if(!Utils.isInteger(e.target.value) && e.target.value != "") {
			state.error = "Promotion Code ID must be an Integer";
		}
		this.setState(state);
	},
	onGenderChange: function(e): any {
		e.preventDefault();
		this.setState({
			gendersSelectedValue: e.target.value,
		});
	},
	onAgeLowerChange: function(e): any {
		e.preventDefault();
		var state = {
			ageLower: e.target.value,
			error: null,
		};
		if(!Utils.isInteger(e.target.value) && e.target.value != "") {
			state.error = "Age Lower must be an Integer"; 
		}
		this.setState(state);
	},
	onAgeUpperChange: function(e): any {
		e.preventDefault();
		var state = {
			ageUpper: e.target.value,
			error: null,
		};
		if(!Utils.isInteger(e.target.value) && e.target.value != "") {
			state.error = "Age Upper must be an Integer"; 
		}
		this.setState(state);
	},
	onStateChange: function(e): any {
		e.preventDefault();
		this.setState({
			state: e.target.value,
		});
	},
	onPharmacyChange: function(e): any {
		e.preventDefault();
		this.setState({
			pharmacy: e.target.value,
		});
	},
	onAdd: function(): any {
		if(this.validateSubmitState()) {
			AdminAPI.addPromotionReferralRoute(
				parseInt(this.state.promotionCodeID), 
				parseInt(this.state.priority), 
				this.state.gendersSelectedValue == "none" ? "" : this.state.gendersSelectedValue, 
				parseInt(this.state.ageLower),
				parseInt(this.state.ageUpper), 
				this.state.state, 
				this.state.pharmacy,
				function(success, data, error){
					if (this.isMounted()) {
						if (!success) {
							this.setState({
								error: error.message
							});
						} else {
							this.setState({
								error: null,
							});
							this.props.onSuccess();
							$("#add-promotion-referral-route-modal").modal('hide');
						}
					}
				}.bind(this));
		}
		return true;
	},
	validateSubmitState: function(): boolean {
		if(!this.state.promotionCodeID) {
			this.setState({error: "promotionCodeID required"});
		} else if (!this.state.priority) {
			this.setState({error: "priority required"});
		} else if (!Utils.isInteger(this.state.promotionCodeID)) {
			this.setState({error: "promotion code id must be an Integer"});
		} else if (this.state.ageLower && !Utils.isInteger(this.state.ageLower)) {
			this.setState({error: "age lower must be an Integer"});
		} else if (this.state.ageUpper && !Utils.isInteger(this.state.ageUpper)) {
			this.setState({error: "age upper must be an Integer"});
		} else {
			return true;
		}
		return false
	},
	render: function(): any {
		return (
			<Modals.ModalForm id="add-promotion-referral-route-modal" title="Add Promotion"
				cancelButtonTitle="Cancel" submitButtonTitle="Add"
				onSubmit={this.onAdd}>

				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}

				<Forms.FormInput label="Promotion Code ID" value={this.state.promotionCodeID} onChange={this.onPromotionCodeIDChange} />
				<Forms.FormInput label="Priority" value={this.state.priority} onChange={this.onPriorityChange} />
				<Forms.FormSelect label="Gender" value={this.state.gendersSelectedValue} opts={this.genders} onChange={this.onGenderChange} />
				<Forms.FormInput label="Age Lower" value={this.state.ageLower} onChange={this.onAgeLowerChange} />
				<Forms.FormInput label="Age Upper" value={this.state.ageUpper} onChange={this.onAgeUpperChange} />
				<Forms.FormInput label="State" value={this.state.state} onChange={this.onStateChange} />
				<Forms.FormInput label="Pharmacy" value={this.state.pharmacy} onChange={this.onPharmacyChange} />
			</Modals.ModalForm>
		);
	}
});

var EditPromotionReferralRouteModal = React.createClass({displayName: "EditPromotionReferralRouteModal",
	// Eventually this should be fetched from the server
	lifecycles: [{name: "Active", value: "ACTIVE"}, {name: "No New Users", value:"NO_NEW_USERS"}, {name: "Deprecated", value: "DEPRECATED"}],
	getInitialState: function(): any {
		return {
			error: null,
			busy: false,
		};
	},
	onLifecycleChange: function(e): any {
		e.preventDefault();
		this.setState({
			lifecyclesSelectedValue: e.target.value
		});
	},
	onUpdate: function(): any {
		AdminAPI.updatePromotionReferralRoute(this.props.referralRoute.id, this.state.lifecyclesSelectedValue, function(success, data, error){
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message
					});
				} else {
					this.setState({
						error: null,
					});
					this.props.onSuccess();
					$("#update-promotion-referral-route-modal").modal('hide');
				}
			}
		}.bind(this));
		return true;
	},
	render: function(): any {
		this.state.lifecyclesSelectedValue = !this.state.lifecyclesSelectedValue ? !this.props.referralRoute ? this.lifecycles[0].value : this.props.referralRoute.lifecycle : this.state.lifecyclesSelectedValue
		return (
			<Modals.ModalForm id="update-promotion-referral-route-modal" title={this.props.referralRoute ? "Update Promotion Referral Route: " + this.props.referralRoute.id : ""}
				cancelButtonTitle="Cancel" submitButtonTitle="Update"
				onSubmit={this.onUpdate}>

				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}

				<Forms.FormSelect label="Lifecycle" value={this.state.lifecyclesSelectedValue} opts={this.lifecycles} onChange={this.onLifecycleChange} />
			</Modals.ModalForm>
		);
	}
});

var PromotionReferralRouteOverview = React.createClass({displayName: "PromotionReferralRouteOverview",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {
			lifecycles: ["ACTIVE", "NO_NEW_USERS"],
		};
	},
	componentWillMount: function() {
		document.title = "Marketing | Promotion Referral Routes";
		this.fetchPromotionReferralRoutes();
	},
	fetchPromotionReferralRoutes: function() {
		AdminAPI.promotionReferralRoutes(this.state.lifecycles, function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message,
						busy: false,
					});
					return;
				}
				this.setState({
					referralRoutes: data.promotion_referral_routes,
					error: null,
					busy: false,
				});
			}
		}.bind(this));
	},
	setRouteForEdit: function(referralRouteForEdit) {
		this.setState({
			referralRouteForEdit: referralRouteForEdit
		})
	},
	toggleDeprecated: function() {
		var i = this.state.lifecycles.indexOf("DEPRECATED")
		var lifecycles = this.state.lifecycles
		if(i === -1){
			lifecycles.push("DEPRECATED")
		} else {
			lifecycles.splice(i, 1)
		}
		this.setState({
			lifecycles: lifecycles,
		});
		this.fetchPromotionReferralRoutes();
	},
	render: function(): any {
		return (
			<div className="container-flow" style={{marginTop: 10}}>
				{Perms.has(Perms.MarketingEdit) ? <AddPromotionReferralRouteModal onSuccess={this.fetchPromotionReferralRoutes} /> : null}
				{Perms.has(Perms.MarketingEdit) ? <EditPromotionReferralRouteModal onSuccess={this.fetchPromotionReferralRoutes} referralRoute={this.state.referralRouteForEdit}/> : null}
				<h2>Promotion Referral Routes</h2>
				{Perms.has(Perms.MarketingEdit) ? <div className="row pull-right"><button className="btn btn-default" data-toggle="modal" data-target="#add-promotion-referral-route-modal">+</button></div> : null}
				{Perms.has(Perms.MarketingView) ? <a href="#" onClick={this.toggleDeprecated}>[Toggle Show Deprecated]</a>: null }
				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				{(Perms.has(Perms.MarketingView) &&	this.state.referralRoutes)? <PromotionReferralRouteList router={this.props.router} referralRoutes={this.state.referralRoutes} setRouteForEdit={this.setRouteForEdit} /> : null}
			</div>
		);
	}
});

var PromotionReferralRouteList = React.createClass({displayName: "PromotionReferralRouteList",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {};
	},
	render: function(): any {
		return (
			<div>
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				{this.props.referralRoutes ? 
					<table className="table">
						<thead>
							<tr>
								<th>Promotion ID</th>
								<th>Lifecycle</th>
								<th>Gender</th>
								<th>Age Lower</th>
								<th>Age Upper</th>
								<th>State</th>
								<th>Pharmacy</th>
								<th>Priority</th>
								<th>Created</th>
								<th>Modified</th>
								{Perms.has(Perms.MarketingEdit) ? <th></th> : null}
							</tr>
						</thead>
						<tbody>
							{this.props.referralRoutes.map(function(rr){return <PromotionReferralRoute key={rr.id} referralRoute={rr} setRouteForEdit={this.props.setRouteForEdit} />}.bind(this))}
						</tbody>
					</table> : null
				}
			</div>
		);
	}
});

var PromotionReferralRoute = React.createClass({displayName: "PromotionReferralRoute",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {};
	},
	onEdit: function(referralRouteForEdit): void {
		this.props.setRouteForEdit(referralRouteForEdit);
		$("#update-promotion-referral-route-modal").modal('show');
	},
	render: function(): any {
		return (
			<tr>
				<td>{this.props.referralRoute.promotion_code_id}</td>
				<td>{this.props.referralRoute.lifecycle}</td>
				<td>{this.props.referralRoute.gender ? this.props.referralRoute.gender : "*"}</td>
				<td>{this.props.referralRoute.age_lower ? this.props.referralRoute.age_lower : "*"}</td>
				<td>{this.props.referralRoute.age_upper ? this.props.referralRoute.age_upper : "*"}</td>
				<td>{this.props.referralRoute.state ? this.props.referralRoute.state : "*"}</td>
				<td>{this.props.referralRoute.pharmacy ? this.props.referralRoute.pharmacy : "*"}</td>
				<td>{this.props.referralRoute.priority}</td>
				<td>{(new Date(this.props.referralRoute.created * 1000)).toString()}</td>
				<td>{(new Date(this.props.referralRoute.modified * 1000)).toString()}</td>
				{Perms.has(Perms.MarketingEdit) ? <td><a href="#" onClick={this.onEdit.bind(this, this.props.referralRoute)}><span className="glyphicon glyphicon-edit" /></a></td> : null}
			</tr>
		);
	}
})
// END: Promotion Referral Route Management

// BEGIN: Promotion Referral Template Management
var SetDefaultPromotionReferralTemplateModal = React.createClass({displayName: "SetDefaultPromotionReferralTemplateModal",
	DefaultStatus: "Default",
	getInitialState: function(): any {
		return {
			error: null,
			busy: false,
		};
	},
	onConfirm: function(): any {
		AdminAPI.updateReferralTemplate(this.props.referralTemplate.id, this.DefaultStatus, function(success, data, error){
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message
					});
				} else {
					this.setState({
						error: null,
					});
					this.props.onSuccess();
					$("#set-promotion-referral-template-default-modal").modal('hide');
				}
			}
		}.bind(this));
		return true;
	},
	render: function(): any {
		return (
			<Modals.ModalForm id="set-promotion-referral-template-default-modal" title={this.props.referralTemplate ? "Set Promotion Referral Template " + this.props.referralTemplate.id + " as Default?" : ""}
				cancelButtonTitle="Cancel" submitButtonTitle="Confirm"
				onSubmit={this.onConfirm}>

				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				<p>This will cause all patients who are not routed to a specific promotion to recieve this promotion.</p>
			</Modals.ModalForm>
		);
	}
});

var SetInactivePromotionReferralTemplateModal = React.createClass({displayName: "SetDefaultPromotionReferralTemplateModal",
	InactiveStatus: "Inactive",
	getInitialState: function(): any {
		return {
			error: null,
			busy: false,
		};
	},
	onConfirm: function(): any {
		AdminAPI.updateReferralTemplate(this.props.referralTemplate.id, this.InactiveStatus, function(success, data, error){
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message
					});
				} else {
					this.setState({
						error: null,
					});
					this.props.onSuccess();
					$("#set-promotion-referral-template-inactive-modal").modal('hide');
				}
			}
		}.bind(this));
		return true;
	},
	render: function(): any {
		return (
			<Modals.ModalForm id="set-promotion-referral-template-inactive-modal" title={this.props.referralTemplate ? "Set Promotion Referral Template " + this.props.referralTemplate.id + " as Inactive?" : ""}
				cancelButtonTitle="Cancel" submitButtonTitle="Confirm"
				onSubmit={this.onConfirm}>

				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				<p>This will cause all patient promotions derived from this template to be inactivated and assigned a new one upon next login.</p>
			</Modals.ModalForm>
		);
	}
});


var AddPromotionReferralTemplateModal = React.createClass({displayName: "AddPromotionReferralTemplateModal",
	getInitialState: function(): any {
		return {
			error: null,
			busy: false,
		};
	},
	onPromotionCodeIDChange: function(e): any {
		e.preventDefault();
		var state = {
			promotionCodeID: e.target.value,
			error: null,
		};
		if(!Utils.isInteger(e.target.value) && e.target.value != "") {
			state.error = "Promotion Code ID must be an Integer"; 
		}
		this.setState(state);
	},
	onTitleChange: function(e): any {
		e.preventDefault();
		this.setState({
			title: e.target.value
		});
	},
	onDescriptionChange: function(e): any {
		e.preventDefault();
		this.setState({
			description: e.target.value
		});
	},
	onImageURLChange: function(e): any {
		e.preventDefault();
		this.setState({
			imageURL: e.target.value,
		});
	},
	onImageWidthChange: function(e): any {
		e.preventDefault();
		var state = {
			imageWidth: e.target.value,
			error: null,
		};
		if(!Utils.isInteger(e.target.value) && e.target.value != "") {
			state.error = "Image Width must be an Integer"; 
		}
		this.setState(state);
	},
	onImageHeightChange: function(e): any {
		e.preventDefault();
		var state = {
			imageHeight: e.target.value,
			error: null,
		};
		if(!Utils.isInteger(e.target.value) && e.target.value != "") {
			state.error = "Image Height must be an Integer"; 
		}
		this.setState(state);
	},
	onDefaultChange: function(e): any {
		e.preventDefault();
		this.setState({
			default: e.target.value
		});
	},
	onFacebookChange: function(e): any {
		e.preventDefault();
		this.setState({
			facebook: e.target.value
		});
	},
	onTwitterChange: function(e): any {
		e.preventDefault();
		this.setState({
			twitter: e.target.value
		});
	},
	onSMSChange: function(e): any {
		e.preventDefault();
		this.setState({
			sms: e.target.value
		});
	},
	onEmailSubjectChange: function(e): any {
		e.preventDefault();
		this.setState({
			email_subject: e.target.value
		});
	},
	onEmailBodyChange: function(e): any {
		e.preventDefault();
		this.setState({
			email_body: e.target.value
		});
	},
	onTextChange: function(e): any {
		e.preventDefault();
		this.setState({
			text: e.target.value
		});
	},
	validateSubmitState: function(): boolean {
		if(!this.state.promotionCodeID) {
			this.setState({error: "promotionCodeID required"});
		} else if (!Utils.isInteger(this.state.promotionCodeID)) {
			this.setState({error: "promotion code id must be an Integer"});
		} else if (!this.state.title) {
			this.setState({error: "title required"});
		} else if (!this.state.description) {
			this.setState({error: "description required"});
		} else if (this.state.imageURL && (!this.state.imageWidth || !this.state.imageHeight)) {
			this.setState({error: "imageWidth and imageHeight required if imageURL present"});
		} else if (this.state.imageWidth && !Utils.isInteger(this.state.imageWidth)) {
			this.setState({error: "imageWidth must be an Integer"});
		} else if (this.state.imageHeight && !Utils.isInteger(this.state.imageHeight)) {
			this.setState({error: "imageHeight must be an Integer"});
		} else if (!this.state.default) {
			this.setState({error: "default required"});
		} else if (!this.state.facebook) {
			this.setState({error: "facebook required"});
		} else if (!this.state.twitter) {
			this.setState({error: "twitter required"});
		} else if (!this.state.sms) {
			this.setState({error: "sms required"});
		} else if (!this.state.email_subject) {
			this.setState({error: "email_subject required"});
		} else if (!this.state.email_body) {
			this.setState({error: "email_body required"});
		} else if (!this.state.text) {
			this.setState({error: "text required"});
		} else {
			return true;
		}
		return false
	},
	onAdd: function(): any {
		if(this.validateSubmitState()){
			AdminAPI.addPromotionReferralTemplate(
				parseInt(this.state.promotionCodeID), 
				this.state.title, 
				this.state.description,
				this.state.imageURL,
				parseInt(this.state.imageWidth),
				parseInt(this.state.imageHeight),
				this.state.default, 
				this.state.facebook, 
				this.state.twitter, 
				this.state.sms, 
				this.state.email_subject, 
				this.state.email_body, 
				this.state.text, 
				function(success, data, error){
				if (this.isMounted()) {
					if (!success) {
						this.setState({
							error: error.message
						});
					} else {
						this.setState({
							error: null,
						});
						this.props.onSuccess();
						$("#add-promotion-referral-template-modal").modal('hide');
					}
				}
			}.bind(this));
		}
		return true;
	},
	render: function(): any {
		return (
			<Modals.ModalForm id="add-promotion-referral-template-modal" title="Add Promotion Referral Template"
				cancelButtonTitle="Cancel" submitButtonTitle="Add"
				onSubmit={this.onAdd}>

				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}

				<Forms.FormInput label="Promotion Code ID" value={this.state.promotionCodeID} onChange={this.onPromotionCodeIDChange} />
				<Forms.FormInput label="Title" value={this.state.title} onChange={this.onTitleChange} />
				<Forms.FormInput label="Description" value={this.state.description} onChange={this.onDescriptionChange} />
				<Forms.FormInput label="*Image URL" value={this.state.imageURL} onChange={this.onImageURLChange} />
				{this.state.imageURL ? <img src={this.state.imageURL} /> : null}
				{this.state.imageURL ? <Forms.FormInput label="Image Width" value={this.state.imageWidth} onChange={this.onImageWidthChange} /> : null}
				{this.state.imageURL ? <Forms.FormInput label="Image Height" value={this.state.imageHeight} onChange={this.onImageHeightChange} /> : null}
				<h4 className="modal-title">Share Text</h4>
				<hr/>
				<Forms.FormInput label="Default" value={this.state.default} onChange={this.onDefaultChange} />
				<Forms.FormInput label="Facebook" value={this.state.facebook} onChange={this.onFacebookChange} />
				<Forms.FormInput label="Twitter" value={this.state.twitter} onChange={this.onTwitterChange} />
				<Forms.FormInput label="SMS" value={this.state.sms} onChange={this.onSMSChange} />
				<Forms.FormInput label="Email Subject" value={this.state.email_subject} onChange={this.onEmailSubjectChange} />
				<Forms.FormInput label="Email Body" value={this.state.email_body} onChange={this.onEmailBodyChange} />
				<h4 className="modal-title">Home Card</h4>
				<hr/>
				<Forms.FormInput label="Text" value={this.state.text} onChange={this.onTextChange} />
			</Modals.ModalForm>
		);
	}
});

var PromotionReferralTemplateOverview = React.createClass({displayName: "PromotionReferralTemplateOverview",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {
			statuses: ["Active","Default"],
		};
	},
	componentWillMount: function() {
		document.title = "Marketing | Promotion Referral Templates";
		this.fetchPromotionReferralTemplates();
	},
	fetchPromotionReferralTemplates: function() {
		AdminAPI.promotionReferralTemplates(this.state.statuses, function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message,
						busy: false,
					});
					return;
				}
				this.setState({
					referralTemplates: data.referral_program_templates,
					error: null,
					busy: false,
				});
			}
		}.bind(this));
	},
	setTemplateForSetDefault: function(referralTemplateForSetDefault) {
		this.setState({
			referralTemplateForSetDefault: referralTemplateForSetDefault
		})
	},
	setTemplateForSetInactive: function(referralTemplateForSetInactive) {
		this.setState({
			referralTemplateForSetInactive: referralTemplateForSetInactive
		})
	},
	render: function(): any {
		return (
			<div className="container-flow" style={{marginTop: 10}}>
				{Perms.has(Perms.MarketingEdit) ? <AddPromotionReferralTemplateModal onSuccess={this.fetchPromotionReferralTemplates} /> : null}
				{Perms.has(Perms.MarketingEdit) ? <SetDefaultPromotionReferralTemplateModal referralTemplate={this.state.referralTemplateForSetDefault} onSuccess={this.fetchPromotionReferralTemplates} /> : null}
				{Perms.has(Perms.MarketingEdit) ? <SetInactivePromotionReferralTemplateModal referralTemplate={this.state.referralTemplateForSetInactive} onSuccess={this.fetchPromotionReferralTemplates} /> : null}
				<h2>Promotion Referral Templates</h2>
				{Perms.has(Perms.MarketingEdit) ? <div className="row pull-right"><button className="btn btn-default" data-toggle="modal" data-target="#add-promotion-referral-template-modal">+</button></div> : null}
				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				{(Perms.has(Perms.MarketingView) &&	this.state.referralTemplates)? <PromotionReferralTemplateList router={this.props.router} referralTemplates={this.state.referralTemplates} setTemplateForSetDefault={this.setTemplateForSetDefault} setTemplateForSetInactive={this.setTemplateForSetInactive}/> : null}
			</div>
		);
	}
});

var PromotionReferralTemplateList = React.createClass({displayName: "PromotionReferralTemplateList",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {};
	},
	render: function(): any {
		return (
			<div>
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				{this.props.referralTemplates ? 
							this.props.referralTemplates.map(function(t){return <PromotionReferralTemplate key={t.id} referralTemplate={t} onSetDefault={this.props.setTemplateForSetDefault} onSetInactive={this.props.setTemplateForSetInactive}/>}.bind(this))
							: null}
			</div>
		);
	}
});

var PromotionReferralTemplate = React.createClass({displayName: "PromotionReferralTemplate",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {};
	},
	onSetDefault: function(referralTemplate): any {
		this.props.onSetDefault(referralTemplate);
		$("#set-promotion-referral-template-default-modal").modal('show');
	},
	onSetInactive: function(referralTemplate): any {
		this.props.onSetInactive(referralTemplate);
		$("#set-promotion-referral-template-inactive-modal").modal('show');
	},
	render: function(): any {
		var showSetDefault = Perms.has(Perms.MarketingEdit) && this.props.referralTemplate.status != "Default"
		var showSetInactive = Perms.has(Perms.MarketingEdit) && this.props.referralTemplate.status != "Default"
		return (
			<div className="card">
				<table className="table">
					<tbody>
						<tr><td>Template ID</td><td>{this.props.referralTemplate.id}</td>
						{ showSetInactive ? <td><div className="pull-right btn btn-default" onClick={this.onSetInactive.bind(this, this.props.referralTemplate)} data-target="#set-promotion-referral-template-inactive-modal">Set Inactive</div></td> : null}
						{ showSetDefault ? <td><div className="pull-right btn btn-default" onClick={this.onSetDefault.bind(this, this.props.referralTemplate)} data-target="#set-promotion-referral-template-default-modal">Make Default</div></td> : null}</tr>
						<tr><td>Status</td><td>{this.props.referralTemplate.status}</td>{ showSetInactive ? <td></td> : null}{ showSetDefault ? <td></td> : null}</tr>
						<tr><td>Promotion Code ID</td><td>{this.props.referralTemplate.promotion_code_id}</td>{ showSetInactive ? <td></td> : null}{ showSetDefault ? <td></td> : null}</tr>
						<tr><td>Created</td><td>{(new Date(this.props.referralTemplate.created * 1000)).toString()}</td>{ showSetInactive ? <td></td> : null}{ showSetDefault ? <td></td> : null}</tr>
						<tr><td>Title</td><td>{this.props.referralTemplate.data.title}</td>{ showSetInactive ? <td></td> : null}{ showSetDefault ? <td></td> : null}</tr>
						<tr><td>Group</td><td>{this.props.referralTemplate.data.group}</td>{ showSetInactive ? <td></td> : null}{ showSetDefault ? <td></td> : null}</tr>
						<tr><td>Description</td><td>{this.props.referralTemplate.data.description}</td>{ showSetDefault ? <td></td> : null}</tr>
						{this.props.referralTemplate.data.image_url ? <tr><td>Image:</td><td><img src={this.props.referralTemplate.data.image_url} /></td><td></td></tr> : null}
						{this.props.referralTemplate.data.image_url ? <tr><td>Image Height:</td><td>{this.props.referralTemplate.data.image_width || "Undefined"} </td><td></td></tr> : null}
						{this.props.referralTemplate.data.image_url ? <tr><td>Image Width:</td><td>{this.props.referralTemplate.data.image_height || "Undefined"} </td><td></td></tr> : null}
					</tbody>
				</table>
				<table className="table">
					<tbody>
						<tr><td>Share Text</td></tr>
						<tr><td></td><td>Default</td><td>{this.props.referralTemplate.data.share_text_params.default}</td></tr>
						<tr><td></td><td>Facebook</td><td>{this.props.referralTemplate.data.share_text_params.facebook}</td></tr>
						<tr><td></td><td>Twitter</td><td>{this.props.referralTemplate.data.share_text_params.twitter}</td></tr>
						<tr><td></td><td>SMS</td><td>{this.props.referralTemplate.data.share_text_params.sms}</td></tr>
						<tr><td></td><td>Email Subject</td><td>{this.props.referralTemplate.data.share_text_params.email_subject}</td></tr>
						<tr><td></td><td>Email Body</td><td>{this.props.referralTemplate.data.share_text_params.email_body}</td></tr>
						<tr><td>Home Card</td></tr>
						<tr><td></td><td>Text</td><td>{this.props.referralTemplate.data.home_card.text}</td></tr>
						<tr><td></td><td>Image URL</td><td>{this.props.referralTemplate.data.home_card.image_url}</td></tr>
					</tbody>
				</table>
			</div>
		);
	}
})
// END: Promotion Referral Template Management


var ActiveFeedbackTemplatesComponent = React.createClass({displayName: "ActiveFeedbackTemplatesComponent",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {
			error: null,
			busy: false,
			templates: []
		}
	},
	componentWillMount: function() {
		document.title = "Marketing | In-app Feedback",
		this.fetchActiveFeedbackTemplates();
	},
	fetchActiveFeedbackTemplates: function() {
		this.setState({
			busy: true, 
			error: null,
		});		
		AdminAPI.getActiveFeedbackTemplates(function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message,
						busy: false,
					})
				} else {
					this.setState({
						error: null,
						templates: data.templates,
						busy: false
					})
				}
			}
		}.bind(this))
	},
	onTemplateAdd: function(e: any) {
		e.preventDefault();
		this.navigate("/marketing/promotions/feedback/template");
	},
	render: function(): any {
		return (
			<div>
				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				<div className="text-right">
					<button className="btn btn-default" onClick={this.onTemplateAdd}>New Feedback Template</button>
				</div>
				<table className="table">
					<thead>
						<tr>
							<th>Tag</th>
							<th>Type</th>
							<th>Created</th>
						</tr>
					</thead>
					<tbody>
						{
							this.state.templates.map(function(t) {
							return (
							<tr key={t.Tag}>
								<td>
									<a href={"feedback/template/"+t.ID} onClick={this.onNavigate}>{t.Tag}</a>
								</td>
								<td>{t.Type}</td>
								<td>{t.Created}</td>
							</tr>
						);
						}.bind(this))}
					</tbody>
				</table>
			</div>
		);
	}
});	

var InAppFeedbackPage = React.createClass({displayName: "InAppFeedbackPage",
	render: function(): any {
		return (
			<div>
				<h2>Rating Level Template Configuration</h2>
				<ConfigureRatingLevelPromptsComponent router={this.props.router} />
				<h2>Active Feedback Templates</h2>
				<ActiveFeedbackTemplatesComponent router={this.props.router} />
			</div>
		);
	}
});

var ConfigureRatingLevelPromptsComponent = React.createClass({displayName: "ConfigureRatingLevelPromptsComponent",
	mixins: [Routing.RouterNavigateMixin],
	getInitialState: function(): any {
		return {
			configs: {}
		};
	},
	componentWillMount: function() {
		document.title = "Marketing | In-app Feedback";
		this.fetchRatingConfigs();
	},
	fetchRatingConfigs: function() {
		this.setState({
			busy: true, 
			error: null,
		});
		AdminAPI.getRatingFeedbackConfig(function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error:error.Message,
						busy: false
					})
					return;
				} 
				this.setState({
					configs: data.configs,
					busy: false,
					error: null,
				})
			}
		}.bind(this));
	},
	onRatingConfigChange: function(rating: number, config: string) {
		var newConfig = this.state.configs;
		newConfig[rating] = config
		this.setState({
			configs: newConfig,
			modified: true
		})
	},
	handleSave: function(e: any) {
		this.setState({
			busy: true, 
			error: null,
		});
		AdminAPI.updateRatingFeedbackConfig(this.state.configs, function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message,
						busy: false,
					})
					return;
				}
				this.setState({
					error: null,
					busy: false,
					modified: false,
				})
				this.fetchRatingConfigs();
			}
		}.bind(this));
	},
	render: function(): any {
		return (
			<div className="container">
				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				<table className="table col-lg-8">
					<thead>
						<tr>
							<th>Rating</th>
							<th>Templates</th>
						</tr>
					</thead>
					<tbody>
						<RatingRow
							rating={1}
							templatesConfig={this.state.configs[1]}
							onRatingConfigChange={this.onRatingConfigChange} />
						<RatingRow
							rating={2}
							templatesConfig = {this.state.configs[2]}
							onRatingConfigChange={this.onRatingConfigChange} />
						<RatingRow
							rating={3}
							templatesConfig = {this.state.configs[3]}
							onRatingConfigChange={this.onRatingConfigChange} />
						<RatingRow
							rating={4}
							templatesConfig = {this.state.configs[4]}
							onRatingConfigChange={this.onRatingConfigChange} />
						<RatingRow
							rating={5}
							templatesConfig = {this.state.configs[5]}
							onRatingConfigChange={this.onRatingConfigChange} />													
					</tbody>
				</table>
				{this.state.modified ?
					<div className="text-right">
						<button
							className = "btn btn-primary"
							onClick={this.handleSave}>Save</button>
					</div>
					: null}
			</div>
		);
	}
});

var RatingRow = React.createClass({displayName:"RatingRow",
	handleChange: function(e: any) {
		e.preventDefault();
		this.props.onRatingConfigChange(this.props.rating, e.target.value);
	},
	render: function(): any {
		return (
			<tr>
				<td>{this.props.rating}</td>
				<td><Forms.FormInput
					value={this.props.templatesConfig}
					onChange = {this.handleChange} />
				</td>	
			</tr>
		);
	}
});


var CreateFeedbackTemplatePage = React.createClass({displayName: "CreateFeedbackTemplatePage",
	mixins: [
		Routing.RouterNavigateMixin,
		React.addons.LinkedStateMixin
	],
	getInitialState: function(): any {
		return {
			selectedTag: "",
			templateTypes: [],
			template: "{}",
			error: null,
			busy: false,
		}
	},
	componentWillMount: function() {
		document.title = "Marketing | In-app Feedback";
		this.fetchFeedbackTemplateTypes();
		if (typeof(this.props.templateID) !== "undefined") {
			this.fetchFeedbackTemplate();
		}
	},
	fetchFeedbackTemplateTypes: function() {
		this.setState({
			busy: true, 
			error: null,
		});
		AdminAPI.getFeedbackTemplateTypes(function(success, data, error) {
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message,
						busy: false,
					});
					return;
				}
				this.setState({
					templateTypes: data.types,
					error: null,
					busy: false,
				});
				this.updateTemplateType(data.types[0].template_type);
			}
		}.bind(this));
	},
	fetchFeedbackTemplate: function() {
		this.setState({
			busy: true, 
			error: null,
		});
		AdminAPI.getFeedbackTemplate(this.props.templateID, function(success, data, error){
			if (this.isMounted()) {
				if (!success) {
					this.setState({
						error: error.message,
						busy: false,
					})
					return;
				} 
				this.setState({
					selectedTag: data.template_data.Tag,
					selectedTemplateType: data.template_data.Type,
					template: JSON.stringify(data.template_data.Template, null, 4)
				})
			} 
		}.bind(this))
	},
	onTemplateSubmit: function(e: any) {
		e.preventDefault();
		this.setState({
			busy: true, 
			error: null,
		});
		AdminAPI.updateFeedbackTemplate(
			this.state.selectedTemplateType, 
			this.state.selectedTag, 
			this.state.template, 
			function(success, data, error) {
				if (this.isMounted()) {
					if (!success) {
						this.setState({
							error: error.message,
							busy: false,
						})
						return;
					}
					this.navigate("/marketing/promotions/feedback");
				}
			}.bind(this));
	},
	onTemplateChange: function(e: any) {
		e.preventDefault();
		var error = null;
		try {
			JSON.parse(e.target.value)
		} catch(ex) {
			error = "Invalid JSON: " + ex.message;
		}

		this.setState({
			error: error,
			template: e.target.value,
		})
	},
	onTemplateTypeChange: function(e: any) {
		e.preventDefault();
		this.updateTemplateType(e.target.value);
	},
	onTemplateTagChange: function(e: any) {
		e.preventDefault();
		this.setState({
			selectedTag: e.target.value,
		})
	},
	updateTemplateType: function(templateType: string) {
		var template = ""
		for (var i =0; i < this.state.templateTypes.length; i++) {
			if (this.state.templateTypes[i].template_type == templateType) {
				template = JSON.stringify(this.state.templateTypes[i].data, null, 4)
			}
		}

		this.setState({
			selectedTemplateType: templateType,
			template: template
		});
	},
	render: function(): any {
		var types = this.state.templateTypes.map(function(t) {
			return {name: t.template_type, value: t.template_type};
		});
		var readOnlyMetadata = false;
		if (this.props.updating) {
			readOnlyMetadata = true;
		}
		return (
			<div>
				{this.state.busy ? <Utils.LoadingAnimation /> : null}
				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				<form role="form" onSubmit={this.onTemplateSubmit} method="PUT">
					<Forms.FormSelect label="Template Type" value={this.state.selectedTemplateType} opts={types} onChange={this.onTemplateTypeChange} disabled={readOnlyMetadata} />
					<Forms.FormInput label="Tag" name="tag" value={this.state.selectedTag} onClick={this.onNavigate} onChange={this.onTemplateTagChange} disabled={readOnlyMetadata} />
					<Forms.TextArea name="template_json" required label="Feedback Template JSON" value={this.state.template} onChange={this.onTemplateChange} rows={20} tabs={true} />
					<div className="text-right">
						<button className="btn btn-primary" type="submit">Save</button>
					</div>
				</form>
			</div>
		);
	}
});


