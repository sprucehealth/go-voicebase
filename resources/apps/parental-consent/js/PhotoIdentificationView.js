/* @flow */

var React = require("react");
var Reflux = require('reflux');
var Utils = require("../../libs/utils.js");
var Constants = require("./Constants.js");
var ParentalConsentActions = require('./ParentalConsentActions.js');
var ParentalConsentStore = require('./ParentalConsentStore.js');
var Spinner = require("spin.js");
var SubmitButtonView = require("./SubmitButtonView.js");

var PhotoIdentificationView = React.createClass({displayName: "PhotoIdentificationView",

	//
	// React
	//
	mixins: [
		Reflux.connect(ParentalConsentStore, 'store'),
		Reflux.listenTo(ParentalConsentActions.uploadGovernmentID, 'governmentIDUploadStarted'),
		Reflux.listenTo(ParentalConsentActions.uploadGovernmentID.completed, 'governmentIDUploadCompleted'),
		Reflux.listenTo(ParentalConsentActions.uploadGovernmentID.failed, 'governmentIDUploadFailed'),
		Reflux.listenTo(ParentalConsentActions.uploadSelfie, 'selfieUploadStarted'),
		Reflux.listenTo(ParentalConsentActions.uploadSelfie.completed, 'selfieUploadCompleted'),
		Reflux.listenTo(ParentalConsentActions.uploadSelfie.failed, 'selfieUploadFailed'),
	],
	propTypes: {
		onFormSubmit: React.PropTypes.func.isRequired,
	},
	getInitialState: function() {
		return {
			isGovernmentIDUploading: false,
			isSelfieUploading: false,
			submitButtonPressedOnce: false,
		}
	},
	componentDidMount: function() {
		// NOTE: to adjust this, go to https://fgnass.github.io/spin.js/, and copy the `opts` var from that page
		var opts = {
		  lines: 11 // The number of lines to draw
		, length: 10 // The length of each line
		, width: 4 // The line thickness
		, radius: 9 // The radius of the inner circle
		, scale: 0.75 // Scales overall size of the spinner
		, corners: 1 // Corner roundness (0..1)
		, color: '#FFF' // #rgb or #rrggbb or array of colors
		, opacity: 0.25 // Opacity of the lines
		, rotate: 0 // The rotation offset
		, direction: 1 // 1: clockwise, -1: counterclockwise
		, speed: 1 // Rounds per second
		, trail: 60 // Afterglow percentage
		, fps: 20 // Frames per second when using setTimeout() as a fallback for CSS
		, zIndex: 2e9 // The z-index (defaults to 2000000000)
		, className: 'spinner' // The CSS class to assign to the spinner
		, top: '50%' // Top position relative to parent
		, left: '50%' // Left position relative to parent
		, shadow: false // Whether to render a shadow
		, hwaccel: false // Whether to use hardware acceleration
		, position: 'absolute' // Element positioning
		}
		var target = document.getElementById('governmentIDSpinner');
		var spinner = new Spinner(opts).spin(target);
		target = document.getElementById('selfieSpinner');
		spinner = new Spinner(opts).spin(target);
	},

	//
	// Action callbacks
	//
	governmentIDUploadStarted: function() {
		this.setState({isGovernmentIDUploading: true})
	},
	governmentIDUploadCompleted: function() {
		this.setState({isGovernmentIDUploading: false})
	},
	governmentIDUploadFailed: function(error: any) {
		alert(error.message)
		// TODO: don't clear out the image if it fails-- instead retry
		this.setState({
			isGovernmentIDUploading: false,
			localGovernmentIDThumbnailSrc: "",
		})
	},
	selfieUploadStarted: function() {
		this.setState({isSelfieUploading: true})
	},
	selfieUploadCompleted: function() {
		this.setState({isSelfieUploading: false})
	},
	selfieUploadFailed: function(error: any) {
		this.setState({isSelfieUploading: false})
	},

	//
	// User interaction callbacks
	// 
	handleSubmit: function(e: any) {
		e.preventDefault();
		this.setState({submitButtonPressedOnce: true})
		if (this.shouldAllowSubmit()) {
			this.props.onFormSubmit({})
		}
	},
	handleGovernmentIDSelection: function(e: any) {
		var fileInput = e.target;
		var files = fileInput.files;
		if (files.length) {
			// Submit to server
			var governmentIDForm = React.findDOMNode(this.refs.governmentIDForm)
			var formData = new FormData(governmentIDForm)
			ParentalConsentActions.uploadGovernmentID(formData);

			// Update the thumbnail
			var file = files[0];
			var imageType = /image.*/; 
			if (!file.type.match(imageType)) {
				console.log("might not be an image");
			}           
			var reader = new FileReader();
			var t = this;
			reader.onload = function(event: any) { 
				var fileReader: FileReader = event.target
				t.setState({localGovernmentIDThumbnailSrc: fileReader.result})
			}
			reader.readAsDataURL(file);
		} else {
			// When the user presses Cancel on that attach file dialog, the files array comes back empty
			// Do nothing, since we don't have a way to delete photos via the API
		}
	},
	handleSelfieSelection: function(e: any) {
		var fileInput = e.target;
		var files = fileInput.files;
		if (files.length) {
			// Submit to server
			var selfieForm = React.findDOMNode(this.refs.selfieForm)
			var formData = new FormData(selfieForm)
			ParentalConsentActions.uploadSelfie(formData);

			// Update the thumbnail
			var file = files[0];
			var imageType = /image.*/;     
			if (!file.type.match(imageType)) {
				console.log("might not be an image");
			}           
			var reader = new FileReader();
			var t = this;
			reader.onload = function(event: any) { 
				var fileReader: FileReader = event.target
				t.setState({localSelfieThumbnailSrc: fileReader.result})
			}
			reader.readAsDataURL(file);
		} else {
			// When the user presses Cancel on that attach file dialog, the files array comes back empty
			// Do nothing, since we don't have a way to delete photos via the API
		}
	},

	//
	// Internal
	//
	shouldAllowSubmit: function(): bool {
		return !this.state.isGovernmentIDUploading 
			&& !this.state.isSelfieUploading 
			&& this.isGovernmentIDFieldValid()
			&& this.isSelfieFieldValid();
	},
	isGovernmentIDFieldValid: function(): bool {
		var store: ParentalConsentStoreType = this.state.store
		return this.state.isGovernmentIDUploading || !Utils.isEmpty(store.identityVerification.serverGovernmentIDThumbnailURL)
	},
	isSelfieFieldValid: function(): bool {
		var store: ParentalConsentStoreType = this.state.store
		return this.state.isSelfieUploading || !Utils.isEmpty(store.identityVerification.serverSelfieThumbnailURL)
	},


	render: function(): any {

		var store: ParentalConsentStoreType = this.state.store

		var uploadFormStyle = {
			width: "100%",
			height: "100%",
		}		
		var fileUploadContainerStyle = {
			width: "100%",
			height: "100%",
			overflow: "hidden",
			position: "absolute",
			top: "0",
			left: "0",
			zIndex: "2",
		}
		var fileUploadInputStyle = {
			display: "block !important",
			width: "100% !important",
			height: "100% !important",
			opacity: "0 !important",
			overflow: "hidden !important",
			cursor: "pointer",
		}
		var uploadContentContainerStyle = {
			alignItems: "center",
			position: "relative",
			zIndex: "1",
		}
		var imageViewContainerStyle = {
			minWidth: "64px",
			marginRight: "16px",
			marginTop: "16px",
			marginBottom: "16px",
		}
		var photoUploadThumbnailStyle = {
			width: "64px",
			height: "64px",
			objectFit: "contain",
		}
		var uploadLabelStyle = {
			marginTop: "auto",
			marginBottom: "auto",
		}

		var placeholderSrc = "http://cl.ly/image/2A1S3t1F0t0j/parental_consent_photo_capture@2x.png"
		
		var governmentIDThumbnailSrc
		if (!Utils.isEmpty(this.state.localGovernmentIDThumbnailSrc)) {
			governmentIDThumbnailSrc = this.state.localGovernmentIDThumbnailSrc
		} else if (!Utils.isEmpty(store.identityVerification.serverGovernmentIDThumbnailURL)) {
			governmentIDThumbnailSrc = store.identityVerification.serverGovernmentIDThumbnailURL
		} else {
			governmentIDThumbnailSrc = placeholderSrc;
		}

		var selfieThumbnailSrc
		if (!Utils.isEmpty(this.state.localSelfieThumbnailSrc)) {
			selfieThumbnailSrc = this.state.localSelfieThumbnailSrc
		} else if (!Utils.isEmpty(store.identityVerification.serverSelfieThumbnailURL)) {
			selfieThumbnailSrc = store.identityVerification.serverSelfieThumbnailURL
		} else {
			selfieThumbnailSrc = placeholderSrc;
		}

		var governmentIDSpinnerContainerDisplay = this.state.isGovernmentIDUploading ? "" : "none"
		var spinnerSpinnerContainerDisplay = this.state.isSelfieUploading ? "" : "none"

		var submitButtonDisabled = !this.shouldAllowSubmit()

		var submitButtonTitle = (this.state.isGovernmentIDUploading || this.state.isSelfieUploading ? "UPLOADING PHOTO..." : "NEXT")

		var governmentIDHighlighted: bool = (this.state.submitButtonPressedOnce ? !this.isGovernmentIDFieldValid() : false)
		var selfieHighlighted: bool = (this.state.submitButtonPressedOnce ? !this.isSelfieFieldValid() : false)

		return (
			<div style={uploadFormStyle}>
				<form encType="multipart/form-data" ref="governmentIDForm">
					<input type="hidden" name="type" value="governmentid" />
					<div className="formFieldRow hasBottomDivider hasTopDivider" style={{marginTop: "20px"}}>
						<div style={uploadContentContainerStyle} className="flexBox justifyContentStartLeft">
							<div style={imageViewContainerStyle}>
								<div style={{
									width: "100%",
									height: "100%",
									backgroundColor: "rgba(0,0,0,0.1)",
									verticalAlign: "middle",
									zIndex: "9999",
									margin: "0px",
									display: governmentIDSpinnerContainerDisplay,
								}}>
									<div style={{
										width: "100%",
										height: "100%",
										backgroundColor: "rgba(0,0,0,0.7)",
										margin: "auto",
										display: "inline-block",
										position: "absolute",
										left: "50%",
										top: "50%",
										transform: "translate(-50%,-50%)",
										WebkitTransform: "translate(-50%,-50%)",
									}}>
										<div id="governmentIDSpinner"></div>
									</div>
								</div>			
								<img src={governmentIDThumbnailSrc} style={photoUploadThumbnailStyle} />
							</div>
							<div style={Utils.mergeProperties(uploadLabelStyle, {
								color: (governmentIDHighlighted ? "#F5A623" : ""),
							})}>
								Take a photo of your government issued photo ID
							</div>
						</div>
						<div style={fileUploadContainerStyle}>
							<input 
								type="file" 
								accept="image/*" 
								onChange={this.handleGovernmentIDSelection} 
								name="file" 
								style={fileUploadInputStyle} 
								disabled={this.state.isGovernmentIDUploading} />
						</div>
					</div>
				</form>	
				<form encType="multipart/form-data" ref="selfieForm">
					<input type="hidden" name="type" value="selfie" />
					<div className="formFieldRow hasBottomDivider">
						<div style={uploadContentContainerStyle} className="flexBox">
							<div style={imageViewContainerStyle}>
								<div style={{
									width: "100%",
									height: "100%",
									backgroundColor: "rgba(0,0,0,0.1)",
									verticalAlign: "middle",
									zIndex: "9999",
									margin: "0px",
									display: spinnerSpinnerContainerDisplay,
								}}>
									<div style={{
										width: "100%",
										height: "100%",
										backgroundColor: "rgba(0,0,0,0.7)",
										margin: "auto",
										display: "inline-block",
										position: "absolute",
										left: "50%",
										top: "50%",
										transform: "translate(-50%,-50%)",
									}}>
										<div id="selfieSpinner"></div>
									</div>
								</div>	
								<img src={selfieThumbnailSrc} style={photoUploadThumbnailStyle} />
							</div>
							<div style={Utils.mergeProperties(uploadLabelStyle, {
								color: (selfieHighlighted ? "#F5A623" : ""),
							})}>
								Take a selfie holding your ID next to your face
							</div>
						</div>
						<div style={fileUploadContainerStyle}>
							<input 
								type="file" 
								accept="image/*" 
								onChange={this.handleSelfieSelection} 
								name="file" 
								style={fileUploadInputStyle} 
								disabled={this.state.isSelfieUploading} />
						</div> 
					</div>
				</form>
				<div>
					<form onSubmit={this.handleSubmit}>
						<SubmitButtonView 
							title={submitButtonTitle}
							appearsDisabled={submitButtonDisabled}/>
					</form>
				</div>
			</div>
		);
	}
});

module.exports = PhotoIdentificationView;