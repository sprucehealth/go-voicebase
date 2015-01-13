/** @jsx React.DOM */

var AdminAPI = require("./api.js");
var Routing = require("../routing.js");
var Utils = require("../utils.js");

module.exports = {
	DrugSearch: React.createClass({displayName: "DrugSearch",
		mixins: [Routing.RouterNavigateMixin],
		getInitialState: function() {
			return {
				query: "",
				busy: false,
				error: null,
				results: null
			};
		},
		componentWillMount: function() {
			document.title = "Search | Drugs | Spruce Admin";
			var q = Utils.getParameterByName("q");
			if (q != this.state.query) {
				this.setState({query: q});
				this.search(q);
			}

			// TODO: need to make sure the page that was navigated to is this one
			// this.navCallback = (function() {
			// 	var q = getParameterByName("q");
			// 	if (q != this.state.query) {
			// 		this.setState({query: q});
			// 		this.search(q);
			// 	}
			// }).bind(this);
			// this.props.router.on("route", this.navCallback);
		},
		componentWillUnmount : function() {
			// this.props.router.off("route", this.navCallback);
		},
		search: function(q) {
			this.props.router.navigate("/drugs?q=" + encodeURIComponent(q), {replace: true}); // TODO: replacing until back tracking works
			if (q == "") {
				this.setState({results: null});
			} else {
				this.setState({busy: true, error: null});
				AdminAPI.searchDrugs(q, function(success, res, error) {
					if (this.isMounted()) {
						if (!success) {
							this.setState({busy: false, results: null, defailts: {}, error: error.message});
							return;
						}
						this.setState({
							busy: false,
							error: null,
							results: res.results || []
						});
					}
				}.bind(this));
			}
		},
		onSearchSubmit: function(e) {
			e.preventDefault();
			this.search(this.state.query);
		},
		onQueryChange: function(e) {
			this.setState({query: e.target.value});
		},
		render: function() {
			return (
				<div className="container doctor-search">
					<div className="row">
						<div className="col-md-3">&nbsp;</div>
						<div className="col-md-6">
							<h2>Search drugs</h2>
							<form onSubmit={this.onSearchSubmit}>
								<div className="form-group">
									<input required autofocus type="text" className="form-control" name="q" value={this.state.query} onChange={this.onQueryChange} />
								</div>
								<button type="submit" className="btn btn-primary btn-lg center-block">Search</button>
							</form>
						</div>
						<div className="col-md-3">&nbsp;</div>
					</div>

					<div className="search-results">
						{this.state.busy ?
							<div className="text-center"><Utils.LoadingAnimation /></div>
						:
							<div>
							{this.state.error ? <div className="text-center"><Utils.Alert type="danger">{this.state.error}</Utils.Alert></div> : null}

							{this.state.results ? DrugSearchResults({
								router: this.props.router,
								results: this.state.results}) : null}
							</div>
						}
					</div>
				</div>
			);
		}
	})
};

var DrugSearchResults = React.createClass({displayName: "DrugSearchResults",
	mixins: [Routing.RouterNavigateMixin],
	render: function() {
		if (this.props.results.length == 0) {
			return (<div className="no-results text-center">No matching drugs found</div>);
		}

		var results = this.props.results.map(function (res) {
			return (
				<div className="row" key={res.name}>
					<div className="col-md-3">&nbsp;</div>
					<div className="col-md-6">
						<DrugSearchResult result={res} router={this.props.router} />
					</div>
					<div className="col-md-3">&nbsp;</div>
				</div>
			);
		}.bind(this))

		return (
			<div>{results}</div>
		);
	}
});

var DrugSearchResult = React.createClass({displayName: "DrugSearchResult",
	mixins: [Routing.RouterNavigateMixin],
	render: function() {
		return (
			<div>
				<strong>{this.props.result.name}</strong><br />
				{this.props.result.error ?
					<div><strong>ERROR: {this.props.result.error}</strong></div>
				:
					<ul>
					{this.props.result.strengths.map(function(st) {
						var ndc = st.medication.RepresentativeNDC;
						return (
							<li key={st.strength}>
								{st.error ?
									<div><strong>ERROR: {st.error}</strong></div>
								:
									<div>
									Generic Product Name: {st.medication.GenericProductName}<br />
									Parsed Generic Name: {st.parsed_generic_name}<br />
									Strength: {st.strength}<br />
									Route: {st.medication.RouteDescription}<br />
									Form: {st.medication.DoseFormDescription}<br />
									Dispense Unit: {st.medication.DispenseUnitDescription}<br />
									OTC: {st.medication.OTC ? "true" : "false"}<br />
									NDC: {ndc}<br />
									{st.guide_id ?
										<a href={"/admin/guides/rx/" + st.guide_id} onClick={this.onNavigate}>RX Guide</a> : null}
									</div>
								}
							</li>
						);
					}.bind(this))}
					</ul>
				}
			</div>
		);
	}
});
