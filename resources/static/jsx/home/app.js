/** @jsx React.DOM */

var Forms = require("../forms.js");
var Utils = require("../utils.js");

var UnsupportedPlatforms = [
	{name: "Select Your Phone", value: ""},
	{name: "Android", value: "Android"},
	{name: "iPhone", value: "iPhone"}
];

var API = {
	// cb is function(success: bool, data: object, jqXHR: jqXHR)
	ajax: function(params, cb) {
		params.success = function(data) {
			cb(true, data, "", null);
		}
		params.error = function(jqXHR) {
			cb(false, null, API.parseError(jqXHR), jqXHR);
		}
		params.url = "/api" + params.url;
		jQuery.ajax(params);
	},
	parseError: function(jqXHR) {
		if (jqXHR.status == 0) {
			return {message: "Network request failed"};
		}
		var err;
		try {
			err = JSON.parse(jqXHR.responseText).error;
		} catch(e) {
			console.error(jqXHR.responseText);
			err = {message: "Unknown error"};
		}
		return err;
	},

	//

	recordForm: function(name, values, cb) {
		this.ajax({
			type: "POST",
			contentType: "application/json",
			url: "/forms/" + encodeURIComponent(name),
			data: JSON.stringify(values),
			dataType: "json"
		}, cb);
	}
};

function staticURL(path) {
	return Spruce.BaseStaticURL + path
}

window.NotifyMeComponent = React.createClass({displayName: "NotifyMeComponent",
	getInitialState: function() {
		return {
			email: "",
			state: "",
			platform: "",
			busy: false,
			error: null
		}
	},
	onSubmit: function(e) {
		e.preventDefault();
		if (this.state.busy) {
			return false;
		}
		this.setState({busy: true, error: null});
		API.recordForm("notify-me", {email: this.state.email, state: this.state.state, platform: this.state.platform},
			function(success, data, error) {
				if (this.isMounted()) {
					if (!success) {
						this.setState({busy: false, error: error});
						return;
					}
					this.setState({busy: false});
					$("#notify-me-modal").modal('hide');
				}
			}.bind(this));
	},
	onChangeEmail: function(e) {
		e.preventDefault();
		this.setState({email: e.target.value});
	},
	onChangeState: function(e) {
		e.preventDefault();
		this.setState({state: e.target.value});
	},
	onChangePlatform: function(e) {
		e.preventDefault();
		this.setState({platform: e.target.value});
	},
	render: function() {
		return (
			<form method="POST" action="#" onSubmit={this.onSubmit} className="text-center">
				<h3>Sign up to be notified when Spruce is available to you.</h3>
				<br />
				<Forms.FormInput placeholder="Your Email Address" value={this.state.email} type="email" required={true} onChange={this.onChangeEmail} />
				<div className="row">
					<div className="col-md-6">
						<Forms.FormSelect value={this.state.state} required={true} onChange={this.onChangeState} opts={Utils.states} />
					</div>
					<div className="col-md-6">
						<Forms.FormSelect value={this.state.platform} required={true} onChange={this.onChangePlatform} opts={UnsupportedPlatforms} />
					</div>
				</div>
				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				<button type="submit" className="btn btn-primary">Sign Up {this.state.busy ? <Utils.LoadingAnimation /> : null}</button>
			</form>
		);
	}
});

window.DoctorInterestComponent = React.createClass({displayName: "DoctorInterestComponent",
	getInitialState: function() {
		return {
			name: "",
			email: "",
			states: "",
			comment: "",
			busy: false,
			error: null
		}
	},
	onSubmit: function(e) {
		e.preventDefault();
		if (this.state.busy) {
			return false;
		}
		this.setState({busy: true, error: null});
		API.recordForm("doctor-interest", {
				name: this.state.name,
				email: this.state.email,
				states: this.state.states,
				comment: this.state.comment
			},
			function(success, data, error) {
				if (this.isMounted()) {
					if (!success) {
						this.setState({busy: false, error: error});
						return;
					}
					this.setState({busy: false});
					$("#doctor-interest-modal").modal('hide');
				}
			}.bind(this));
	},
	onChangeName: function(e) {
		e.preventDefault();
		this.setState({name: e.target.value});
	},
	onChangeEmail: function(e) {
		e.preventDefault();
		this.setState({email: e.target.value});
	},
	onChangeStates: function(e) {
		e.preventDefault();
		this.setState({states: e.target.value});
	},
	onChangeComment: function(e) {
		e.preventDefault();
		this.setState({comment: e.target.value});
	},
	render: function() {
		return (
			<form method="POST" action="#" onSubmit={this.onSubmit} className="text-center">
				<h2>Get In Touch</h2>
				<p>Tell us a little bit about yourself and someone from Spruce will be in touch shortly.</p>
				<Forms.FormInput placeholder="Your Name" value={this.state.name} required={true} onChange={this.onChangeName} />
				<Forms.FormInput placeholder="Your Email Address" value={this.state.email} type="email" required={true} onChange={this.onChangeEmail} />
				<Forms.FormInput placeholder="States Where You're Licensed" value={this.state.states} required={true} onChange={this.onChangeStates} />
				<Forms.FormInput placeholder="Optional Comment" value={this.state.comment} onChange={this.onChangeComment} />
				{this.state.error ? <Utils.Alert type="danger">{this.state.error}</Utils.Alert> : null}
				<button type="submit" className="btn btn-primary">Submit {this.state.busy ? <Utils.LoadingAnimation /> : null}</button>
			</form>
		);
	}
});

window.PressComponent = React.createClass({displayName: "PressComponent",
	press: [
		{name: "Harpers Bazaar", image: "press_bazaar_promo.png", quote: '"#1 must-have beauty app."'},
		{name: "Wired", image: "press_wired_promo.png", quote: '"As a patient, you can start to take care of your skin problem in a few minutes, instead of scheduling an appointment and waiting a few weeks."'},
		{name: "SELF", image: "press_self_promo_cropped.png", quote: '"Whatever the source of your breakouts, you\'re now carrying the solution in your purse."'},
		{name: "Gizmodo", image: "press_gizmodo_promo.png", quote: '"Spruce brings acne sufferers the future of telemedicine."'},
		{name: "Cosmopolitan", image: "press_cosmo_promo.png", quote: '"There\'s An App for Acne. And it Works."'},
	],
	getInitialState: function() {
		return {active: 0};
	},
	handlePressSwitch: function(e) {
		e.preventDefault();
		var index = parseInt(e.target.dataset.index);
		if (!isNaN(index) && index != this.state.active) {
			this.setState({active: index});
		}
	},
	render: function() {
		return (
			<div>
				<div className="quote">{this.press[this.state.active].quote}</div>

				{this.press.map(function(p, i) {
					return (
						<span key={"press-"+i}>
							<img
								src={staticURL("/img/press/" + p.image)}
								alt={p.name}
								className={i==this.state.active?"active":""}
								onMouseOver={this.handlePressSwitch}
								onClick={this.handlePressSwitch}
								data-index={i} />
							{" "}
						</span>
					);
				}.bind(this))}
			</div>
		);
	}
});

