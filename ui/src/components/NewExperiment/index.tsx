import { Box, Button, Container, Drawer, Fab, useMediaQuery, useTheme } from '@material-ui/core'
import { Experiment, StepperFormProps } from './types'
import { Form, Formik, FormikHelpers } from 'formik'
import React, { useState } from 'react'
import { StepperProvider, useStepperContext } from './Context'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { defaultExperimentSchema, validationSchema } from './constants'
import { setAlert, setAlertOpen, setNeedToRefreshExperiments } from 'slices/globalStatus'

import AddIcon from '@material-ui/icons/Add'
import CancelIcon from '@material-ui/icons/Cancel'
import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import PublishIcon from '@material-ui/icons/Publish'
import Stepper from './Stepper'
import api from 'api'
import { parseSubmitValues } from 'lib/formikhelpers'
import { useHistory } from 'react-router-dom'
import { useStoreDispatch } from 'store'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    container: {
      display: 'flex',
      flexDirection: 'column',
      width: '50vw',
      height: '100%',
      padding: theme.spacing(6),
      [theme.breakpoints.down('md')]: {
        width: '100vw',
      },
    },
    new: {
      [theme.breakpoints.down('xs')]: {
        display: 'none',
      },
    },
    fab: {
      [theme.breakpoints.up('sm')]: {
        display: 'none',
      },
      [theme.breakpoints.down('xs')]: {
        position: 'fixed',
        bottom: theme.spacing(7.5),
        right: theme.spacing(3),
        display: 'flex',
        background: '#fff',
        color: theme.palette.primary.main,
        zIndex: 1101,
      },
    },
  })
)

interface ActionsProps {
  isSubmitting?: boolean
  toggleDrawer: () => void
}

const Actions = ({ isSubmitting = false, toggleDrawer }: ActionsProps) => {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const size = isTabletScreen ? ('small' as 'small') : ('medium' as 'medium')

  const { state } = useStepperContext()

  return (
    <Box display="flex" justifyContent="space-between" marginBottom={6}>
      <Button variant="outlined" size={size} startIcon={<CancelIcon />} onClick={toggleDrawer}>
        Cancel
      </Button>
      <Box display="flex">
        <Box mr={3}>
          <Button variant="outlined" size={size} startIcon={<CloudUploadOutlinedIcon />}>
            Yaml File
          </Button>
        </Box>
        <Button
          type="submit"
          variant="contained"
          size={size}
          color="primary"
          startIcon={<PublishIcon />}
          disabled={state.activeStep < 4 || isSubmitting}
        >
          Submit
        </Button>
      </Box>
    </Box>
  )
}

export default function NewExperiment() {
  const initialValues: Experiment = defaultExperimentSchema

  const classes = useStyles()
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const history = useHistory()
  const dispatch = useStoreDispatch()

  const [open, setOpen] = useState(false)
  const toggleDrawer = () => setOpen(!open)

  const handleOnSubmit = (values: Experiment, actions: FormikHelpers<Experiment>) => {
    const parsedValues = parseSubmitValues(values)

    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug parsedValues:', parsedValues)
    }

    api.experiments
      .newExperiment(parsedValues)
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: 'Created successfully!',
          })
        )
        dispatch(setAlertOpen(true))

        if (history.location.pathname === '/experiments') {
          dispatch(setNeedToRefreshExperiments(true))
        } else {
          history.push('/experiments')
        }
      })
      .catch(console.log)
      .finally(toggleDrawer)
  }

  return (
    <>
      <Button
        className={classes.new}
        variant="outlined"
        size={isTabletScreen ? ('small' as 'small') : ('medium' as 'medium')}
        startIcon={<AddIcon />}
        onClick={toggleDrawer}
      >
        New Experiment
      </Button>
      <Fab className={classes.fab} color="inherit" size="medium" aria-label="New experiment">
        <AddIcon onClick={toggleDrawer} />
      </Fab>
      <Drawer anchor="right" open={open} onClose={toggleDrawer}>
        <StepperProvider>
          <Formik initialValues={initialValues} validationSchema={validationSchema} onSubmit={handleOnSubmit}>
            {(props: StepperFormProps) => {
              const { isSubmitting } = props

              return (
                <Container className={classes.container}>
                  <Form>
                    <Actions isSubmitting={isSubmitting} toggleDrawer={toggleDrawer} />
                    <Stepper formProps={props} />
                  </Form>
                </Container>
              )
            }}
          </Formik>
        </StepperProvider>
      </Drawer>
    </>
  )
}
