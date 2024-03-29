export const bind: (
  keys: Array<string> | string,
  cb: (e: {stopPropagation: () => void}, key: string) => void,
  type: 'keydown'
) => void
export const unbind: (keys: Array<string> | string) => void
