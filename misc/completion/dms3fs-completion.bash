_dms3fs_comp()
{
    COMPREPLY=( $(compgen -W "$1" -- ${word}) )
    if [[ ${#COMPREPLY[@]} == 1 && ${COMPREPLY[0]} == "--"*"=" ]] ; then
        # If there's only one option, with =, then discard space
        compopt -o nospace
    fi
}

_dms3fs_help_only()
{
    _dms3fs_comp "--help"
}

_dms3fs_add()
{
    if [[ "${prev}" == "--chunker" ]] ; then
        _dms3fs_comp "placeholder1 placeholder2 placeholder3" # TODO: a) Give real options, b) Solve autocomplete bug for "="
    elif [ "${prev}" == "--pin" ] ; then
        _dms3fs_comp "true false"
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--recursive --quiet --silent --progress --trickle --only-hash --wrap-with-directory --hidden --chunker= --pin= --raw-leaves --help "
    else
        _dms3fs_filesystem_complete
    fi
}

_dms3fs_bitswap()
{
    dms3fs_comp "ledger stat unwant wantlist --help"
}

_dms3fs_bitswap_ledger()
{
    _dms3fs_help_only
}

_dms3fs_bitswap_stat()
{
    _dms3fs_help_only
}

_dms3fs_bitswap_unwant()
{
    _dms3fs_help_only
}

_dms3fs_bitswap_wantlist()
{
    dms3fs_comp "--peer= --help"
}

_dms3fs_bitswap_unwant()
{
    _dms3fs_help_only
}

_dms3fs_block()
{
    _dms3fs_comp "get put rm stat --help"
}

_dms3fs_block_get()
{
    _dms3fs_hash_complete
}

_dms3fs_block_put()
{
    if [ "${prev}" == "--format" ] ; then
        _dms3fs_comp "v0 placeholder2 placeholder3" # TODO: a) Give real options, b) Solve autocomplete bug for "="
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--format= --help"
    else
        _dms3fs_filesystem_complete
    fi
}

_dms3fs_block_rm()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--force --quiet --help"
    else
        _dms3fs_hash_complete
    fi
}

_dms3fs_block_stat()
{
    _dms3fs_hash_complete
}

_dms3fs_bootstrap()
{
    _dms3fs_comp "add list rm --help"
}

_dms3fs_bootstrap_add()
{
    _dms3fs_comp "default --help"
}

_dms3fs_bootstrap_list()
{
    _dms3fs_help_only
}

_dms3fs_bootstrap_rm()
{
    _dms3fs_comp "all --help"
}

_dms3fs_cat()
{
    if [[ ${prev} == */* ]] ; then
        COMPREPLY=() # Only one argument allowed
    elif [[ ${word} == */* ]] ; then
        _dms3fs_hash_complete
    else
        _dms3fs_pinned_complete
    fi
}

_dms3fs_commands()
{
    _dms3fs_comp "--flags --help"
}

_dms3fs_config()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--bool --json"
    elif [[ ${prev} == *.* ]] ; then
        COMPREPLY=() # Only one subheader of the config can be shown or edited.
    else
        _dms3fs_comp "show edit replace"
    fi
}

_dms3fs_config_edit()
{
    _dms3fs_help_only
}

_dms3fs_config_replace()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--help"
    else
        _dms3fs_filesystem_complete
    fi
}

_dms3fs_config_show()
{
    _dms3fs_help_only
}

_dms3fs_daemon()
{
    if [[ ${prev} == "--routing" ]] ; then
        _dms3fs_comp "dht dhtclient none" # TODO: Solve autocomplete bug for "="
    elif [[ ${prev} == "--mount-dms3fs" ]] || [[ ${prev} == "--mount-dms3ns" ]] || [[ ${prev} == "=" ]]; then
        _dms3fs_filesystem_complete
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--init --routing= --mount --writable --mount-dms3fs= \
            --mount-dms3ns= --unrestricted-api --disable-transport-encryption \
            -- enable-gc --manage-fdlimit --offline --migrate --help"
    fi
}

_dms3fs_dag()
{
    _dms3fs_comp "get put --help"
}

_dms3fs_dag_get()
{
    _dms3fs_help_only
}

_dms3fs_dag_put()
{
    if [[ ${prev} == "--format" ]] ; then
        _dms3fs_comp "cbor placeholder1" # TODO: a) Which format more then cbor is valid? b) Solve autocomplete bug for "="
    elif [[ ${prev} == "--input-enc" ]] ; then
        _dms3fs_comp "json placeholder1" # TODO: a) Which format more then json is valid? b) Solve autocomplete bug for "="
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--format= --input-enc= --help"
    else
        _dms3fs_filesystem_complete
    fi
}

_dms3fs_dht()
{
    _dms3fs_comp "findpeer findprovs get provide put query --help"
}

_dms3fs_dht_findpeer()
{
    _dms3fs_comp "--verbose --help"
}

_dms3fs_dht_findprovs()
{
    _dms3fs_comp "--verbose --help"
}

_dms3fs_dht_get()
{
    _dms3fs_comp "--verbose --help"
}

_dms3fs_dht_provide()
{
    _dms3fs_comp "--recursive --verbose --help"
}

_dms3fs_dht_put()
{
    _dms3fs_comp "--verbose --help"
}

_dms3fs_dht_query()
{
    _dms3fs_comp "--verbose --help"
}

_dms3fs_diag()
{
    _dms3fs_comp "sys cmds net --help"
}

_dms3fs_diag_cmds()
{
    if [[ ${prev} == "clear" ]] ; then
        return 0
    elif [[ ${prev} =~ ^-?[0-9]+$ ]] ; then
        _dms3fs_comp "ns us µs ms s m h" # TODO: Trigger with out space, eg. "dms3fs diag set-time 10ns" not "... set-time 10 ns"
    elif [[ ${prev} == "set-time" ]] ; then
        _dms3fs_help_only
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--verbose --help"
    else
        _dms3fs_comp "clear set-time"
    fi
}

_dms3fs_diag_sys()
{
    _dms3fs_help_only
}

_dms3fs_diag_net()
{
    if [[ ${prev} == "--vis" ]] ; then
        _dms3fs_comp "d3 dot text" # TODO: Solve autocomplete bug for "="
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--timeout= --vis= --help"
    fi
}

_dms3fs_dns()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--recursive --help"
    fi
}

_dms3fs_files()
{
    _dms3fs_comp "mv rm flush read write cp ls mkdir stat"
}

_dms3fs_files_mv()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--recursive --flush"
    elif [[ ${word} == /* ]] ; then
        _dms3fs_files_complete
    else
        COMPREPLY=( / )
        [[ $COMPREPLY = */ ]] && compopt -o nospace
    fi
}

_dms3fs_files_rm()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--recursive --flush"
    elif [[ ${word} == /* ]] ; then
        _dms3fs_files_complete
    else
        COMPREPLY=( / )
        [[ $COMPREPLY = */ ]] && compopt -o nospace
    fi
}
_dms3fs_files_flush()
{
    if [[ ${word} == /* ]] ; then
        _dms3fs_files_complete
    else
        COMPREPLY=( / )
        [[ $COMPREPLY = */ ]] && compopt -o nospace
    fi
}

_dms3fs_files_read()
{
    if [[ ${prev} == "--count" ]] || [[ ${prev} == "--offset" ]] ; then
        COMPREPLY=() # Numbers, just keep it empty
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--offset --count --help"
    elif [[ ${word} == /* ]] ; then
        _dms3fs_files_complete
    else
        COMPREPLY=( / )
        [[ $COMPREPLY = */ ]] && compopt -o nospace
    fi
}

_dms3fs_files_write()
{
    if [[ ${prev} == "--count" ]] || [[ ${prev} == "--offset" ]] ; then # Dirty check
        COMPREPLY=() # Numbers, just keep it empty
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--offset --count --create --truncate --help"
    elif [[ ${prev} == /* ]] ; then
        _dms3fs_filesystem_complete
    elif [[ ${word} == /* ]] ; then
        _dms3fs_files_complete
    else
        COMPREPLY=( / )
        [[ $COMPREPLY = */ ]] && compopt -o nospace
    fi
}

_dms3fs_files_cp()
{
    if [[ ${word} == /* ]] ; then
        _dms3fs_files_complete
    else
        COMPREPLY=( / )
        [[ $COMPREPLY = */ ]] && compopt -o nospace
    fi
}

_dms3fs_files_ls()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "-l --help"
    elif [[ ${prev} == /* ]] ; then
        COMPREPLY=() # Path exist
    elif [[ ${word} == /* ]] ; then
        _dms3fs_files_complete
    else
        COMPREPLY=( / )
        [[ $COMPREPLY = */ ]] && compopt -o nospace
    fi
}

_dms3fs_files_mkdir()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--parents --help"

    elif [[ ${prev} == /* ]] ; then
        COMPREPLY=() # Path exist
    elif [[ ${word} == /* ]] ; then
        _dms3fs_files_complete
    else
        COMPREPLY=( / )
        [[ $COMPREPLY = */ ]] && compopt -o nospace
    fi
}

_dms3fs_files_stat()
{
    if [[ ${prev} == /* ]] ; then
        COMPREPLY=() # Path exist
    elif [[ ${word} == /* ]] ; then
        _dms3fs_files_complete
    else
        COMPREPLY=( / )
        [[ $COMPREPLY = */ ]] && compopt -o nospace
    fi
}

_dms3fs_file()
{
    if [[ ${prev} == "ls" ]] ; then
        _dms3fs_hash_complete
    else
        _dms3fs_comp "ls --help"
    fi
}

_dms3fs_file_ls()
{
    _dms3fs_help_only
}

_dms3fs_get()
{
    if [ "${prev}" == "--output" ] ; then
        compopt -o default # Re-enable default file read
        COMPREPLY=()
    elif [ "${prev}" == "--compression-level" ] ; then
        _dms3fs_comp "-1 1 2 3 4 5 6 7 8 9" # TODO: Solve autocomplete bug for "="
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--output= --archive --compress --compression-level= --help"
    else
        _dms3fs_hash_complete
    fi
}

_dms3fs_id()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--format= --help"
    fi
}

_dms3fs_init()
{
    _dms3fs_comp "--bits --force --empty-repo --help"
}

_dms3fs_log()
{
    _dms3fs_comp "level ls tail --help"
}

_dms3fs_log_level()
{
    # TODO: auto-complete subsystem and level
    _dms3fs_help_only
}

_dms3fs_log_ls()
{
    _dms3fs_help_only
}

_dms3fs_log_tail()
{
    _dms3fs_help_only
}

_dms3fs_ls()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--headers --resolve-type=false --help"
    else
        _dms3fs_hash_complete
    fi
}

_dms3fs_mount()
{
    if [[ ${prev} == "--dms3fs-path" ]] || [[ ${prev} == "--dms3ns-path" ]] || [[ ${prev} == "=" ]] ; then
        _dms3fs_filesystem_complete
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--dms3fs-path= --dms3ns-path= --help"
    fi
}

_dms3fs_name()
{
    _dms3fs_comp "publish resolve --help"
}

_dms3fs_name_publish()
{
    if [[ ${prev} == "--lifetime" ]] || [[ ${prev} == "--ttl" ]] ; then
        COMPREPLY=() # Accept only numbers
    elif [[ ${prev} =~ ^-?[0-9]+$ ]] ; then
        _dms3fs_comp "ns us µs ms s m h" # TODO: Trigger without space, eg. "dms3fs diag set-time 10ns" not "... set-time 10 ns"
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--resolve --lifetime --ttl --help"
    elif [[ ${word} == */ ]]; then
        _dms3fs_hash_complete
    else
        _dms3fs_pinned_complete
    fi
}

_dms3fs_name_resolve()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--recursive --nocache --help"
    fi
}

_dms3fs_object()
{
    _dms3fs_comp "data diff get links new patch put stat --help"
}

_dms3fs_object_data()
{
    _dms3fs_hash_complete
}

_dms3fs_object_diff()
{
  if [[ ${word} == -* ]] ; then
      _dms3fs_comp "--verbose --help"
  else
      _dms3fs_hash_complete
  fi
}


_dms3fs_object_get()
{
    if [ "${prev}" == "--encoding" ] ; then
        _dms3fs_comp "protobuf json xml"
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--encoding --help"
    else
        _dms3fs_hash_complete
    fi
}

_dms3fs_object_links()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--headers --help"
    else
        _dms3fs_hash_complete
    fi
}

_dms3fs_object_new()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--help"
    else
        _dms3fs_comp "unixfs-dir"
    fi
}

_dms3fs_object_patch()
{
    if [[ -n "${COMP_WORDS[3]}" ]] ; then # Root merkledag object exist
        case "${COMP_WORDS[4]}" in
        append-data)
            _dms3fs_help_only
            ;;
        add-link)
            if [[ ${word} == -* ]] && [[ ${prev} == "add-link" ]] ; then # Dirty check
                _dms3fs_comp "--create"
            #else
                # TODO: Hash path autocomplete. This is tricky, can be hash or a name.
            fi
            ;;
        rm-link)
            _dms3fs_hash_complete
            ;;
        set-data)
            _dms3fs_filesystem_complete
            ;;
        *)
            _dms3fs_comp "append-data add-link rm-link set-data"
            ;;
        esac
    else
        _dms3fs_hash_complete
    fi
}

_dms3fs_object_put()
{
    if [ "${prev}" == "--inputenc" ] ; then
        _dms3fs_comp "protobuf json"
    elif [ "${prev}" == "--datafieldenc" ] ; then
        _dms3fs_comp "text base64"
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--inputenc --datafieldenc --help"
    else
        _dms3fs_hash_complete
    fi
}

_dms3fs_object_stat()
{
    _dms3fs_hash_complete
}

_dms3fs_pin()
{
    _dms3fs_comp "rm ls add --help"
}

_dms3fs_pin_add()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--recursive=  --help"
    elif [[ ${word} == */ ]] && [[ ${word} != "/dms3fs/" ]] ; then
        _dms3fs_hash_complete
    fi
}

_dms3fs_pin_ls()
{
    if [[ ${prev} == "--type" ]] || [[ ${prev} == "-t" ]] ; then
        _dms3fs_comp "direct indirect recursive all" # TODO: Solve autocomplete bug for
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--count --quiet --type= --help"
    elif [[ ${word} == */ ]] && [[ ${word} != "/dms3fs/" ]] ; then
        _dms3fs_hash_complete
    fi
}

_dms3fs_pin_rm()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--recursive  --help"
    elif [[ ${word} == */ ]] && [[ ${word} != "/dms3fs/" ]] ; then
        COMPREPLY=() # TODO: _dms3fs_hash_complete() + List local pinned hashes as default?
    fi
}

_dms3fs_ping()
{
    _dms3fs_comp "--count=  --help"
}

_dms3fs_pubsub()
{
    _dms3fs_comp "ls peers pub sub --help"
}

_dms3fs_pubsub_ls()
{
    _dms3fs_help_only
}

_dms3fs_pubsub_peers()
{
    _dms3fs_help_only
}

_dms3fs_pubsub_pub()
{
    _dms3fs_help_only
}

_dms3fs_pubsub_sub()
{
    _dms3fs_comp "--discover --help"
}

_dms3fs_refs()
{
    if [ "${prev}" == "--format" ] ; then
        _dms3fs_comp "src dst linkname"
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "local --format= --edges --unique --recursive --help"
    #else
        # TODO: Use "dms3fs ref" and combine it with autocomplete, see _dms3fs_hash_complete
    fi
}

_dms3fs_refs_local()
{
    _dms3fs_help_only
}

_dms3fs_repo()
{
    _dms3fs_comp "fsck gc stat verify version --help"
}

_dms3fs_repo_version()
{
    _dms3fs_comp "--quiet --help"
}

_dms3fs_repo_verify()
{
    _dms3fs_help_only
}

_dms3fs_repo_gc()
{
    _dms3fs_comp "--quiet --help"
}

_dms3fs_repo_stat()
{
    _dms3fs_comp "--human --help"
}

_dms3fs_repo_fsck()
{
    _dms3fs_help_only
}

_dms3fs_resolve()
{
    if [[ ${word} == /dms3fs/* ]] ; then
        _dms3fs_hash_complete
    elif [[ ${word} == /dms3ns/* ]] ; then
        COMPREPLY=() # Can't autocomplete dms3ns
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--recursive --help"
    else
        opts="/dms3ns/ /dms3fs/"
        COMPREPLY=( $(compgen -W "${opts}" -- ${word}) )
        [[ $COMPREPLY = */ ]] && compopt -o nospace
    fi
}

_dms3fs_stats()
{
    _dms3fs_comp "bitswap bw repo --help"
}

_dms3fs_stats_bitswap()
{
    _dms3fs_help_only
}

_dms3fs_stats_bw()
{
    # TODO: Which protocol is valid?
    _dms3fs_comp "--peer= --proto= --poll --interval= --help"
}

_dms3fs_stats_repo()
{
    _dms3fs_comp "--human= --help"
}

_dms3fs_swarm()
{
    _dms3fs_comp "addrs connect disconnect filters peers --help"
}

_dms3fs_swarm_addrs()
{
    _dms3fs_comp "local --help"
}

_dms3fs_swarm_addrs_local()
{
    _dms3fs_comp "--id --help"
}

_dms3fs_swarm_connect()
{
    _dms3fs_multiaddr_complete
}

_dms3fs_swarm_disconnect()
{
    local OLDIFS="$IFS" ; local IFS=$'\n' # Change divider for iterator one line below
    opts=$(for x in `dms3fs swarm peers`; do echo ${x} ; done)
    IFS="$OLDIFS" # Reset divider to space, ' '
    COMPREPLY=( $(compgen -W "${opts}" -- ${word}) )
    [[ $COMPREPLY = */ ]] && compopt -o nospace -o filenames
}

_dms3fs_swarm_filters()
{
    if [[ ${prev} == "add" ]] || [[ ${prev} == "rm" ]]; then
        _dms3fs_multiaddr_complete
    else
        _dms3fs_comp "add rm --help"
    fi
}

_dms3fs_swarm_filters_add()
{
    _dms3fs_help_only
}

_dms3fs_swarm_filters_rm()
{
    _dms3fs_help_only
}

_dms3fs_swarm_peers()
{
    _dms3fs_help_only
}

_dms3fs_tar()
{
    _dms3fs_comp "add cat --help"
}

_dms3fs_tar_add()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--help"
    else
        _dms3fs_filesystem_complete
    fi
}

_dms3fs_tar_cat()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--help"
    else
        _dms3fs_filesystem_complete
    fi
}

_dms3fs_update()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--version" # TODO: How does "--verbose" option work?
    else
        _dms3fs_comp "versions version install stash revert fetch"
    fi
}

_dms3fs_update_install()
{
    if   [[ ${prev} == v*.*.* ]] ; then
        COMPREPLY=()
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--version"
    else
        local OLDIFS="$IFS" ; local IFS=$'\n' # Change divider for iterator one line below
        opts=$(for x in `dms3fs update versions`; do echo ${x} ; done)
        IFS="$OLDIFS" # Reset divider to space, ' '
        COMPREPLY=( $(compgen -W "${opts}" -- ${word}) )
    fi
}

_dms3fs_update_stash()
{
    if [[ ${word} == -* ]] ; then
        _dms3fs_comp "--tag --help"
    fi
}
_dms3fs_update_fetch()
{
    if [[ ${prev} == "--output" ]] ; then
        _dms3fs_filesystem_complete
    elif [[ ${word} == -* ]] ; then
        _dms3fs_comp "--output --help"
    fi
}

_dms3fs_version()
{
    _dms3fs_comp "--number --commit --repo"
}

_dms3fs_hash_complete()
{
    local lastDir=${word%/*}/
    echo "LastDir: ${lastDir}" >> ~/Downloads/debug-dms3fs.txt
    local OLDIFS="$IFS" ; local IFS=$'\n' # Change divider for iterator one line below
    opts=$(for x in `dms3fs file ls ${lastDir}`; do echo ${lastDir}${x}/ ; done) # TODO: Implement "dms3fs file ls -F" to get rid of frontslash after files. This take long time to run first time on a new shell.
    echo "Options: ${opts}" >> ~/Downloads/debug-dms3fs.txt
    IFS="$OLDIFS" # Reset divider to space, ' '
    echo "Current: ${word}" >> ~/Downloads/debug-dms3fs.txt
    COMPREPLY=( $(compgen -W "${opts}" -- ${word}) )
    echo "Suggestion: ${COMPREPLY}" >> ~/Downloads/debug-dms3fs.txt
    [[ $COMPREPLY = */ ]] && compopt -o nospace -o filenames # Removing whitespace after output & handle output as filenames. (Only printing the latest folder of files.)
    return 0
}

_dms3fs_files_complete()
{
    local lastDir=${word%/*}/
    local OLDIFS="$IFS" ; local IFS=$'\n' # Change divider for iterator one line below
    opts=$(for x in `dms3fs files ls ${lastDir}`; do echo ${lastDir}${x}/ ; done) # TODO: Implement "dms3fs files ls -F" to get rid of frontslash after files. This does currently throw "Error: /cats/foo/ is not a directory"
    IFS="$OLDIFS" # Reset divider to space, ' '
    COMPREPLY=( $(compgen -W "${opts}" -- ${word}) )
    [[ $COMPREPLY = */ ]] && compopt -o nospace -o filenames
    return 0
}

_dms3fs_multiaddr_complete()
{
    local lastDir=${word%/*}/
    # Special case
    if [[ ${word} == */"ipcidr"* ]] ; then # TODO: Broken, fix it.
        opts="1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32" # TODO: IPv6?
        COMPREPLY=( $(compgen -W "${opts}" -- ${word}) )
    # "Loop"
    elif [[ ${word} == /*/ ]] || [[ ${word} == /*/* ]] ; then
        if [[ ${word} == /*/*/*/*/*/ ]] ; then
            COMPREPLY=()
        elif [[ ${word} == /*/*/*/*/ ]] ; then
            word=${word##*/}
            opts="dms3fs/ "
            COMPREPLY=( $(compgen -W "${opts}" -- ${word}) )
        elif [[ ${word} == /*/*/*/ ]] ; then
            word=${word##*/}
            opts="4001/ "
            COMPREPLY=( $(compgen -W "${opts}" -- ${word}) )
        elif [[ ${word} == /*/*/ ]] ; then
            word=${word##*/}
            opts="udp/ tcp/ ipcidr/"
            COMPREPLY=( $(compgen -W "${opts}" -- ${word}) )
        elif [[ ${word} == /*/ ]] ; then
            COMPREPLY=() # TODO: This need to return something to NOT break the function. Maybe a "/" in the end as well due to -o filename option.
        fi
        COMPREPLY=${lastDir}${COMPREPLY}
    else # start case
        opts="/ip4/ /ip6/"
        COMPREPLY=( $(compgen -W "${opts}" -- ${word}) )
    fi
    [[ $COMPREPLY = */ ]] && compopt -o nospace -o filenames
    return 0
}

_dms3fs_pinned_complete()
{
    local OLDIFS="$IFS" ; local IFS=$'\n'
    local pinned=$(dms3fs pin ls)
    COMPREPLY=( $(compgen -W "${pinned}" -- ${word}) )
    IFS="$OLDIFS"
    if [[ ${#COMPREPLY[*]} -eq 1 ]]; then # Only one completion, remove pretty output
        COMPREPLY=( ${COMPREPLY[0]/ *//} ) #Remove ' ' and everything after
        [[ $COMPREPLY = */ ]] && compopt -o nospace  # Removing whitespace after output
    fi
}
_dms3fs_filesystem_complete()
{
    compopt -o default # Re-enable default file read
    COMPREPLY=()
}

_dms3fs()
{
    COMPREPLY=()
    compopt +o default # Disable default to not deny completion, see: http://stackoverflow.com/a/19062943/1216348

    local word="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"

    case "${COMP_CWORD}" in
        1)
            local opts="add bitswap block bootstrap cat commands config daemon dag dht \
                        diag dns file files get id init log ls mount name object pin ping pubsub \
                        refs repo resolve stats swarm tar update version"
            COMPREPLY=( $(compgen -W "${opts}" -- ${word}) );;
        2)
            local command="${COMP_WORDS[1]}"
            eval "_dms3fs_$command" 2> /dev/null ;;
        *)
            local command="${COMP_WORDS[1]}"
            local subcommand="${COMP_WORDS[2]}"
            eval "_dms3fs_${command}_${subcommand}" 2> /dev/null && return
            eval "_dms3fs_$command" 2> /dev/null ;;
    esac
}
complete -F _dms3fs dms3fs
