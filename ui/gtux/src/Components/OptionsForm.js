import React from 'react';
import TrackOptions from "./TrackOptionsSelect";
import BaseOptions from "./BaseOptionsSelect";

export default class OptionsForm extends React.Component {
	state = {
		'binSize': 500000,
	}

	optionsUpdate = (group,value) => {
		let options = this.props.options;

		let binSize = this.state.binSize;
		switch (group) {
			case 'left':
				options.left = value;
				options.left.bin_size = binSize;
				break;
			case 'right':
				options.right = value;
				options.right.bin_size = binSize;
				break;
			case 'general':
				binSize = value.binSize;
				options.general.title = value.title;
				options.general.tick_interval = value.rulerInterval;
				options.general.display_ruler = value.rulerDisplay.value;
				options.left.bin_size = binSize;
				options.right.bin_size = binSize;
				this.setState({binSize});
				break;
		}
		this.props.setOptions(options);
	}

	render(props,state) {
		const { genotypes } = this.props;
		return (
			<fieldset className={'genotype-field'} >
				<legend>Options </legend>
				<BaseOptions optionsUpdate={(group,value)=>this.optionsUpdate(group,value)} />
				<TrackOptions side={'Left'} genotypes={genotypes} optionsUpdate={(group,value)=>this.optionsUpdate(group,value)}/>
				<TrackOptions side={'Right'} genotypes={genotypes} optionsUpdate={(group,value)=>this.optionsUpdate(group,value)}/>
			</fieldset>
		);
	}
}