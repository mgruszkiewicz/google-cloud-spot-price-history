TMPFOLDER="/tmp/pricing-data"
mkdir -p $TMPFOLDER
for rev in $(git log --pretty=format:"%h" "${1}")
do
	echo $rev
    datestamp="$(git show --no-patch --no-notes --date=format:'%Y-%m-%d.%H%M.%S' --pretty=format:"%ad" "${rev}")"
    filestamp="${datestamp}.${rev}"
    git show "${rev}":"${1}" > "${TMPFOLDER}/${filestamp}"
done

