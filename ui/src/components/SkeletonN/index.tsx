import React from 'react'
import Skeleton from '@material-ui/lab/Skeleton'

const SkeletonN: React.FC<{ n: number }> = ({ n }) => (
  <>
    {n > 0 &&
      Array(n)
        .fill(0)
        .map((_, i) => <Skeleton key={i} />)}
  </>
)

export default SkeletonN
