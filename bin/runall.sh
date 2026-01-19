#!/usr/bin/env bash
#############################################################################
#
set -u
#############################################################################
#
USERDATA=/Users/wraith/JetBrains/tribenet/ottomap/testdata
OTTOMAP_ROOT=/Users/wraith/JetBrains/tribenet/ottomap
OTTOMAP="${OTTOMAP_ROOT}/build/ottomap"

#############################################################################
#
YEARS=(0899 0900 0901 0902 0903 0904)
MONTHS=(01 02 03 04 05 06 07 08 09 10 11 12)
CLANS=(0346) #  0500 0501 0987

#############################################################################
# build the ottomap executable
build_ottomap() {
  echo " info: building local executable..."
  cd "${OTTOMAP_ROOT}" || return 2
  go build -o "${OTTOMAP}" || { echo "error: unable to build local executable"; return 2; }
  "${OTTOMAP}" version || return 2
}

#############################################################################
# Render a single map.
#
# Returns:
#   0 = success
#   1 = render failed (soft error; stop this clan)
#   2 = hard error    (abort script)
render_map() {
  local clan="$1"
  local year="$2"
  local month="$3"

  if [[ "${clan}" == "0346" && "${year}-${month}" < "0903-06" ]]; then
    [[ "${DEBUG:-}" ]] && echo " debug: skipping ${clan} ${year}-${month} (known bad input)"
    return 0
  fi

  local reportExtractFile="data/input/${year}-${month}.${clan}.report.txt"
  [ -f "${reportExtractFile}" ] || return 0

  local mapRenderFile="data/output/${year}-${month}.${clan}.render.wxx"
  local mapRenderLog="data/logs/${year}-${month}.${clan}.render.log"
  local mapRenderErrorLog="data/errors/${year}-${month}.${clan}.render.err"

  echo " info: rendering ${clan}/${mapRenderFile}"

  if ! "${OTTOMAP}" render --show-version --quiet \
        --clan-id "${clan}" --max-turn "${year}-${month}" \
        --auto-eol --save-with-turn-id --show-grid-coords --shift-map \
        2> "${mapRenderLog}"
  then
    mv "${mapRenderLog}" "${mapRenderErrorLog}" || return 2
    echo " warn: clan ${clan}: render failed at ${year}-${month}"
    echo "       see ${mapRenderErrorLog}"
    return 1
  fi

  return 0
}

#############################################################################
#
render_clan() {
  local clan="$1"
  cd "${USERDATA}/${clan}" || return 2

  mkdir -p data/{errors,input,logs,output}
  rm -f data/{errors,logs}/*
  rm -f data/output/*."${clan}".wxx

  local year month
  for year in "${YEARS[@]}"; do
    for month in "${MONTHS[@]}"; do
      render_map "${clan}" "${year}" "${month}"
      case $? in
        0) continue ;;   # success or no input
        1) return 0 ;;   # soft failure: stop this clan
        *) return 2 ;;   # hard failure
      esac
    done
  done

  echo " info: clan ${clan}: complete"
  return 0
}

#############################################################################
#
build_ottomap || exit $?

#############################################################################
#
for clan in "${CLANS[@]}"; do
  render_clan "${clan}" || exit $?
done

exit 0
