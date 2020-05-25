import { FormikProps } from 'formik'

export interface ExperimentBasic {
  name: string
  namespace: string
  labels?: { [key: string]: string }
}

export interface ExperimentScope {
  namespaceSelector: string[]
  phaseSelector: string[]
  mode: string
  value: string
  labelSelector?: { [key: string]: string }
}

export interface ExperimentTargetPod {
  action: string
  container: string
}

export interface ExperimentNetworkDelay {
  latency: string
  correlation: string
  jitter: string
}

export interface ExperimentTargetNetwork {
  action: string
  delay: ExperimentNetworkDelay
}

export interface ExperimentTarget {
  pod: ExperimentTargetPod
  network: ExperimentTargetNetwork
}

export interface ExperimentSchedule {
  cron: string
  duration: string
}

export interface Experiment {
  basic: ExperimentBasic
  scope: ExperimentScope
  target: ExperimentTarget
  schedule: ExperimentSchedule
}

export interface StepperState {
  activeStep: number
}

export type StepperAction = { type: 'next' } | { type: 'back' } | { type: 'jump'; step: number } | { type: 'reset' }

export type StepperDispatch = (action: StepperAction) => void

export interface StepperContextProps {
  state: StepperState
  dispatch: StepperDispatch
}

export type StepperFormProps = FormikProps<Experiment>

export interface StepProps {
  formProps: StepperFormProps
}
