import React from 'react'
import { withStyles } from '@material-ui/core/styles'
import MuiTooltip from '@material-ui/core/Tooltip'

const styles = theme => ({
  lightTooltip: {
    background: theme.palette.primary.contrastText,
    color: theme.palette.text.primary,
    boxShadow: theme.shadows[24],
    ...theme.typography.h6
  }
})

const Tooltip = ({ title, children, classes }) => {
  return (
    <MuiTooltip title={title} classes={{ tooltip: classes.lightTooltip }}>
      {children}
    </MuiTooltip>
  )
}

export default withStyles(styles)(Tooltip)
