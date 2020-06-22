import { RouteProps } from 'react-router'

import Overview from './pages/Overview'
import Experiments from './pages/Experiments'
import ExperimentDetail from './pages/ExperimentDetail'
import Events from './pages/Events'
import Archives from './pages/Archives'

const routes: RouteProps[] = [
  {
    component: Overview,
    path: '/overview',
    exact: true,
  },
  {
    component: Experiments,
    path: '/experiments',
    exact: true,
  },
  {
    component: ExperimentDetail,
    path: '/experiments/:name',
  },
  {
    component: Events,
    path: '/events',
    exact: true,
  },
  {
    component: Archives,
    path: '/archives',
    exact: true,
  },
]

export default routes
