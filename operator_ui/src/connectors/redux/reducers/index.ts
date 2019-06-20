import { combineReducers } from 'redux'
import accountBalances from './accountBalances'
import authentication from './authentication'
import bridges from './bridges'
import configuration from './configuration'
import create from './create'
import dashboardIndex, { IState as IDashboardState } from './dashboardIndex'
import fetching from './fetching'
import jobRuns from './jobRuns'
import jobs from './jobs'
import notifications from './notifications'
import redirect from './redirect'
import transactions from './transactions'
import transactionsIndex from './transactionsIndex'

export interface IState {
  dashboardIndex: IDashboardState
}

const reducer = combineReducers({
  accountBalances,
  authentication,
  bridges,
  configuration,
  create,
  dashboardIndex,
  fetching,
  jobRuns,
  jobs,
  notifications,
  redirect,
  transactions,
  transactionsIndex
})

export default reducer
