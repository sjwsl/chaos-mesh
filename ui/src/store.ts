import { AnyAction, configureStore, getDefaultMiddleware } from '@reduxjs/toolkit'

import rootReducer from 'reducers'
import { useDispatch } from 'react-redux'

export type RootState = ReturnType<typeof rootReducer>

const middlewares = [...getDefaultMiddleware()]

const genStore = () => {
  if (process.env.NODE_ENV === 'development') {
    const { createLogger } = require('redux-logger')

    const logger = createLogger({
      predicate: (_: any, action: AnyAction) =>
        !['experiments/state/pending', 'experiments/state/fulfilled'].includes(action.type),
    })

    middlewares.push(logger)
  }

  const store = configureStore({
    reducer: rootReducer,
    middleware: middlewares,
    devTools: process.env.NODE_ENV !== 'production',
  })

  return store
}

const store = genStore()

type StoreDispatch = typeof store.dispatch
export const useStoreDispatch = () => useDispatch<StoreDispatch>()

export default store
