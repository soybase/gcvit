import React from 'react';

export default class ComparisonGenotype extends React.Component {
    state= {hideInfo:true};
    render () {
        const {hideInfo} = this.state;
        return (
            <div>
                <div className={'row'}>
                    <div
                        className={'u-full-width fake-button'}
                        onClick={()=>{this.setState({hideInfo:!hideInfo})}}
                    >
                        <div className={'row'}>
                            <div className={'one column'} > <div className={hideInfo ? 'arrow-down rotate':'arrow-down'}/> </div>
                            <div className={'ten columns'}> Comparison Genotype </div>
                            <div className={'one column'} > <div className={hideInfo ? 'arrow-down rotate':'arrow-down'}/> </div>
                        </div>
                    </div>
                </div>
                <div className={'row'}>
                    <div className={'help-text row'} style={{maxHeight: hideInfo ? '0px' : '50%'}}>
                        <div className={'u-full-width modal-section'}>
                            <p>
                                <ul>
                                    <li> Pick a Color that will represent the reference genome. Please note that this color will only show up when “total” is indicated in the display options. </li>
                                    <li> At this time, comparisons may only be made within the dataset of the selected reference. </li>
                                    <li> Select which genotype will be the reference by either typing in the name of the genotype or using the pull down menu to see available genotypes. For comparisons, each selected genotyped with be compared against this reference.</li>
                                    <li> Remove a comparison by clicking on the X </li>
                                </ul>
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}