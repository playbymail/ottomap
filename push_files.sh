#!/bin/bash
RSYNC_PROGRESS=--progress

# confirm that we're running from the root of the repository
[ -d build ] || {
  echo error: must run from the root of the repository
  exit 2
}

# build the local executable, get the version number, then remove it
LOCAL_EXE=build/ottomap
echo " info: building local executable..."
go build -o "${LOCAL_EXE}" || {
  echo "error: unable to build local executable"
  exit 2
}
OTTOVER=$( "${LOCAL_EXE}" version )
if [ -z "${OTTOVER}" ]; then
  echo "error: '${LOCAL_EXE} version' seems to have failed"
  exit 2
fi
rm -f "${LOCAL_EXE}"

echo " info: building executables for version '${OTTOVER}'"

echo " info: building linux executable..."
LINUX_EXE=build/ottomap
GOOS=linux GOARCH=amd64 go build -o "${LINUX_EXE}" || exit 2

echo " info: building windows executable..."
WINDOWS_EXE="build/ottomap-windows-${OTTOVER}.exe"
GOOS=windows GOARCH=amd64 go build -o "${WINDOWS_EXE}" || exit 2

# push the executable files to our production server
echo " info: pushing executable files to mdhender/bin..."
rsync -av ${RSYNC_PROGRESS} "${LINUX_EXE}"   mdhender@tribenet:"bin/" || {
  echo "error: failed to copy the linux executable to the production server"
  exit 2
}
rsync -av ${RSYNC_PROGRESS} "${LINUX_EXE}"   mdhender@tribenet:"bin/ottomap.${OTTOVER}" || {
  echo "error: failed to copy the linux executable to the production server"
  exit 2
}

echo " info: pushing linux executable files to ottomap.mdhenderson.com/bin..."
rsync -av ${RSYNC_PROGRESS} "${LINUX_EXE}"   mdhender@tribenet:"/var/www/ottomap.mdhenderson.com/bin/" || {
  echo "error: failed to copy the linux executable to the production server"
  exit 2
}
rsync -av ${RSYNC_PROGRESS} "${LINUX_EXE}"   mdhender@tribenet:"/var/www/ottomap.mdhenderson.com/bin/ottomap.${OTTOVER}" || {
  echo "error: failed to copy the linux executable to the production server"
  exit 2
}

echo " info: pushing windows executable files to ottomap.mdhenderson.com/assets/uploads..."
rsync -av ${RSYNC_PROGRESS} "${WINDOWS_EXE}" mdhender@tribenet:"/var/www/ottomap.mdhenderson.com/assets/uploads/" || {
  echo "error: failed to copy the windows executable to the production server"
  exit 2
}

echo " info: removing build files..."
rm -f "${LINUX_EXE}" "${WINDOWS_EXE}"

echo " info: push to production server succeeded"

exit 0
