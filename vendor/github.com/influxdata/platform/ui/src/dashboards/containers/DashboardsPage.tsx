// Libraries
import React, {PureComponent} from 'react'
import {InjectedRouter} from 'react-router'
import {connect} from 'react-redux'
import download from 'src/external/download'
import _ from 'lodash'

// Components
import DashboardsContents from 'src/dashboards/components/DashboardsPageContents'
import {Page} from 'src/pageLayout'
import SearchWidget from 'src/shared/components/search_widget/SearchWidget'
import {
  OverlayTechnology,
  Button,
  ComponentColor,
  IconFont,
} from 'src/clockface'
import ImportDashboardOverlay from 'src/dashboards/components/ImportDashboardOverlay'

// Utils
import {getDeep} from 'src/utils/wrappers'

// APIs
import {createDashboard} from 'src/dashboards/apis/v2'

// Actions
import {
  getDashboardsAsync,
  importDashboardAsync,
  deleteDashboardAsync,
} from 'src/dashboards/actions/v2'
import {setDefaultDashboard} from 'src/shared/actions/links'
import {retainRangesDashTimeV1 as retainRangesDashTimeV1Action} from 'src/dashboards/actions/v2/ranges'
import {notify as notifyAction} from 'src/shared/actions/notifications'

import {
  dashboardSetDefaultFailed,
  dashboardExported,
  dashboardExportFailed,
  dashboardCreateFailed,
} from 'src/shared/copy/notifications'

// Types
import {Notification} from 'src/types/notifications'
import {DashboardFile, Cell} from 'src/types/v2/dashboards'
import {Links, Dashboard} from 'src/types/v2'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

interface Props {
  router: InjectedRouter
  links: Links
  handleSetDefaultDashboard: typeof setDefaultDashboard
  handleGetDashboards: typeof getDashboardsAsync
  handleDeleteDashboard: typeof deleteDashboardAsync
  handleImportDashboard: typeof importDashboardAsync
  notify: (message: Notification) => void
  retainRangesDashTimeV1: (dashboardIDs: string[]) => void
  dashboards: Dashboard[]
}

interface State {
  searchTerm: string
  isImportingDashboard: boolean
}

@ErrorHandling
class DashboardsPage extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props)

    this.state = {
      searchTerm: '',
      isImportingDashboard: false,
    }
  }

  public async componentDidMount() {
    const {handleGetDashboards, dashboards, links} = this.props
    await handleGetDashboards(links.dashboards)
    const dashboardIDs = dashboards.map(d => d.id)
    this.props.retainRangesDashTimeV1(dashboardIDs)
  }

  public render() {
    const {dashboards, notify, links} = this.props
    const {searchTerm} = this.state

    return (
      <>
        <Page>
          <Page.Header fullWidth={false}>
            <Page.Header.Left>
              <Page.Title title="Dashboards" />
            </Page.Header.Left>
            <Page.Header.Right>
              <SearchWidget
                placeholderText="Filter dashboards by name..."
                onSearch={this.filterDashboards}
              />
              <Button
                onClick={this.handleToggleOverlay}
                icon={IconFont.Import}
                text="Import Dashboard"
                titleText="Import a dashboard from a file"
              />
              <Button
                color={ComponentColor.Primary}
                onClick={this.handleCreateDashboard}
                icon={IconFont.Plus}
                text="Create Dashboard"
                titleText="Create a new dashboard"
              />
            </Page.Header.Right>
          </Page.Header>
          <Page.Contents fullWidth={false} scrollable={true}>
            <DashboardsContents
              dashboards={dashboards}
              onSetDefaultDashboard={this.handleSetDefaultDashboard}
              defaultDashboardLink={links.defaultDashboard}
              onDeleteDashboard={this.handleDeleteDashboard}
              onCreateDashboard={this.handleCreateDashboard}
              onCloneDashboard={this.handleCloneDashboard}
              onExportDashboard={this.handleExportDashboard}
              notify={notify}
              searchTerm={searchTerm}
            />
          </Page.Contents>
        </Page>
        {this.renderImportOverlay}
      </>
    )
  }

  private handleSetDefaultDashboard = async (
    defaultDashboardLink: string
  ): Promise<void> => {
    const {dashboards, notify, handleSetDefaultDashboard} = this.props
    const {name} = dashboards.find(d => d.links.self === defaultDashboardLink)

    try {
      await handleSetDefaultDashboard(defaultDashboardLink)
    } catch (error) {
      console.error(error)
      notify(dashboardSetDefaultFailed(name))
    }
  }

  private handleCreateDashboard = async (): Promise<void> => {
    const {links, router, notify} = this.props
    try {
      const newDashboard = {
        name: 'Name this dashboard',
        cells: [],
      }
      const data = await createDashboard(links.dashboards, newDashboard)
      router.push(`/dashboards/${data.id}`)
    } catch (error) {
      notify(dashboardCreateFailed())
    }
  }

  private handleCloneDashboard = (dashboard: Dashboard) => async (): Promise<
    void
  > => {
    const {router, links, notify} = this.props
    const name = `${dashboard.name} (clone)`
    try {
      const data = await createDashboard(links.dashboards, {
        ...dashboard,
        name,
      })
      router.push(`/dashboards/${data.id}`)
    } catch (error) {
      notify(dashboardCreateFailed())
    }
  }

  private handleDeleteDashboard = (dashboard: Dashboard) => (): void => {
    this.props.handleDeleteDashboard(dashboard)
  }

  private handleExportDashboard = (dashboard: Dashboard) => async (): Promise<
    void
  > => {
    const dashboardForDownload = await this.modifyDashboardForDownload(
      dashboard
    )
    try {
      download(
        JSON.stringify(dashboardForDownload, null, '\t'),
        `${dashboard.name}.json`,
        'text/plain'
      )
      this.props.notify(dashboardExported(dashboard.name))
    } catch (error) {
      this.props.notify(dashboardExportFailed(dashboard.name, error))
    }
  }

  private modifyDashboardForDownload = async (
    dashboard: Dashboard
  ): Promise<DashboardFile> => {
    return {meta: {chronografVersion: '2.0'}, dashboard}
  }

  private handleImportDashboard = async (
    dashboard: Dashboard
  ): Promise<void> => {
    const defaultCell = {
      x: 0,
      y: 0,
      w: 4,
      h: 4,
    }

    const {links} = this.props
    const name = _.get(dashboard, 'name', 'Name this dashboard')
    const cellsWithDefaultsApplied = getDeep<Cell[]>(
      dashboard,
      'cells',
      []
    ).map(c => ({...defaultCell, ...c}))

    await this.props.handleImportDashboard(links.dashboards, {
      ...dashboard,
      name,
      cells: cellsWithDefaultsApplied,
    })
  }

  private filterDashboards = (searchTerm: string): void => {
    this.setState({searchTerm})
  }

  private handleToggleOverlay = (): void => {
    this.setState({isImportingDashboard: !this.state.isImportingDashboard})
  }

  private get renderImportOverlay(): JSX.Element {
    const {notify} = this.props
    const {isImportingDashboard} = this.state

    return (
      <OverlayTechnology visible={isImportingDashboard}>
        <ImportDashboardOverlay
          onDismissOverlay={this.handleToggleOverlay}
          onImportDashboard={this.handleImportDashboard}
          notify={notify}
        />
      </OverlayTechnology>
    )
  }
}

const mstp = state => {
  const {dashboards, links} = state

  return {
    dashboards,
    links,
  }
}

const mdtp = {
  notify: notifyAction,
  handleSetDefaultDashboard: setDefaultDashboard,
  handleGetDashboards: getDashboardsAsync,
  handleDeleteDashboard: deleteDashboardAsync,
  handleImportDashboard: importDashboardAsync,
  retainRangesDashTimeV1: retainRangesDashTimeV1Action,
}

export default connect(
  mstp,
  mdtp
)(DashboardsPage)
