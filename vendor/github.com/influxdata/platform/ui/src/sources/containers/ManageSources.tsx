import React, {PureComponent} from 'react'
import {connect} from 'react-redux'
import {ErrorHandling} from 'src/shared/decorators/errors'

import * as sourcesActions from 'src/shared/actions/sources'
import {notify as notifyAction} from 'src/shared/actions/notifications'

import {Page} from 'src/pageLayout'
import InfluxTable from 'src/sources/components/InfluxTable'

import {sourceDeleted, sourceDeleteFailed} from 'src/shared/copy/notifications'

import {Source} from 'src/types/v2'
import {Notification, Service} from 'src/types'

interface Props {
  source: Source
  sources: Source[]
  services: Service[]
  notify: (n: Notification) => void
  removeAndLoadSources: sourcesActions.RemoveAndLoadSources
}

const VERSION = process.env.npm_package_version

@ErrorHandling
class ManageSources extends PureComponent<Props> {
  public render() {
    const {sources, source} = this.props

    return (
      <Page>
        <Page.Header fullWidth={false}>
          <Page.Header.Left>
            <Page.Title title="Configuration" />
          </Page.Header.Left>
          <Page.Header.Right />
        </Page.Header>
        <Page.Contents fullWidth={false} scrollable={true}>
          <InfluxTable
            source={source}
            sources={sources}
            onDeleteSource={this.handleDeleteSource}
          />
          <p className="version-number">Chronograf Version: {VERSION}</p>
        </Page.Contents>
      </Page>
    )
  }

  private handleDeleteSource = (source: Source) => {
    const {notify} = this.props

    try {
      this.props.removeAndLoadSources(source)
      notify(sourceDeleted(source.name))
    } catch (e) {
      notify(sourceDeleteFailed(source.name))
    }
  }
}

const mstp = ({source, sources, services}) => ({
  source,
  sources,
  services,
})

const mdtp = {
  notify: notifyAction,
  removeAndLoadSources: sourcesActions.removeAndLoadSources,
}

export default connect(mstp, mdtp)(ManageSources)
