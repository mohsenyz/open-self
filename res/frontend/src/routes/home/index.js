import {Component, h} from 'preact';
import style from './style';
import {getAppsUsage} from "../../utils/api";

export default class Home extends Component{

	state = {
		appsUsage : []
	};

	getAppsUsage() {
		getAppsUsage().then((appsUsage) => {
			this.setState({appsUsage : appsUsage})
		})
	}

	componentDidMount() {
		this.getAppsUsage()
	}

	render() {
		return (
			<div className={style.home}>
				<h1>Home</h1>
			</div>
		)
	}
};
