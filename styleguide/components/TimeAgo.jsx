import React from 'react'
import TimeAgoNoTooltip from 'react-time-ago/no-tooltip'
import Tooltip from './Tooltip'

const TimeAgo = ({ children }) => (
  <Tooltip title={children}>
    <span>
      <TimeAgoNoTooltip tooltip={false}>
        {Date.parse(children)}
      </TimeAgoNoTooltip>
    </span>
  </Tooltip>
)

export default TimeAgo
