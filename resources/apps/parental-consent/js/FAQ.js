/* @flow */

var React = require('react/addons');
var Utils = require("../../libs/utils.js");

var Analytics = require("../../libs/analytics.js");
var AnalyticsScreenName = "faq"
var Constants = require("./Constants.js");

var ReactCSSTransitionGroup = React.addons.CSSTransitionGroup;

module.exports = {
	Component: React.createClass({displayName: "FAQComponent",
		componentDidMount: function() {
			Analytics.record(AnalyticsScreenName + "_viewed", {"app_type": Constants.AnalyticsAppType, "screen_id": AnalyticsScreenName})
		},
		render: function(): any {
			return (
				<div>
					{this.props.faq.Sections.map(function(s) {
						return <FAQSection key={s.Title} title={s.Title} questions={s.Questions} />
					})}
				</div>
			);
		}
	})
};

var FAQSection = React.createClass({displayName: "FAQSection",
	render: function(): any {
		return (
			<div className="section">
				<h3>{this.props.title}</h3>
				{this.props.questions.map(function(q) {
					return (
						<div key={q.Question}>
							<hr />
							<FAQEntry question={q.Question} answer={q.Answer} />
						</div>
					);
				})}
				<hr />
			</div>
		);
	}
});

var FAQEntry = React.createClass({displayName: "FAQEntry",
	getInitialState: function() {
		return {expanded: false};
	},
	handleExpand: function(e) {
		e.preventDefault();
		this.setState({expanded: !this.state.expanded});
	},
	render: function(): any {
		return (
			<div className="qa">
				<div className="question">
					<span>
						<a href="#" onClick={this.handleExpand}>{this.props.question}</a>
					</span>
					<img
						src={this.state.expanded ?
							Utils.staticURL("/img/faq/faq_circle_minus@2x.png") :
							Utils.staticURL("/img/faq/faq_circle_plus@2x.png")}
						width="27"
						height="27"
						onClick={this.handleExpand} />
				</div>
				<ReactCSSTransitionGroup transitionName="faq-answer">
					{this.state.expanded ?
						<div key={this.props.question} className="answer" dangerouslySetInnerHTML={{__html: this.props.answer}} />
					: null }
				</ReactCSSTransitionGroup>
			</div>
		);
	}
});
