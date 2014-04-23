#! /bin/sh

### BEGIN INIT INFO
# Provides:          brog
# Required-Start:    $local_fs $remote_fs $network $syslog
# Required-Stop:     $local_fs $remote_fs $network $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Brog blogging platform
# Description:       Brog blogging platform
### END INIT INFO



# Script variable names should be lower-case not to conflict with internal
# /bin/sh variables such as PATH, EDITOR or SHELL.
app_root="/home/brog" #Adapt in fuction of your actual install
app_user="brog"
pid_path="$app_root"
socket_path="$app_root"
brog_pid_path="$pid_path/brog.pid"
brog_socket_path="$socket_path/brog.sock"


### Here ends user configuration ###


# Switch to the app_user if it is not them who is running the script.
if [ "$(whoami)" != "$app_user" ]; then
  sudo -u "$app_user" -H -i $0 "$@"; exit;
fi

# Switch to the Brog path, if it fails exit with an error.
if ! cd "$app_root" ; then
 echo "Failed to cd into '$app_root', exiting!";  exit 1
fi

### Init Script functions

check_pids(){
  if ! [ -d "$pid_path" ]; then
    echo "Could not find the path '$pid_path' needed to store the pids."
    echo "Check your configuration"
    exit 1
  fi
  # If there exists a file which should hold the value of the Brog pid: read it.
  if [ -f "$brog_pid_path" ]; then
    wpid=$(cat "$brog_pid_path")
  else
    wpid=0
  fi
}

# We use the pids in so many parts of the script it makes sense to always check them.
# Only after start() is run should the pids change.
check_pids


# Checks whether the different parts of the service are already running or not.
check_status(){
  check_pids
  # If brog is running kill -0 $wpid returns true, or rather 0.
  # Checks of brog_status should only check for == 0 or != 0, never anything else.
  if [ $wpid -ne 0 ]; then
    kill -0 "$wpid" 2>/dev/null
    brog_status="$?"
  else
    brog_status="-1"
  fi
}

# Check for stale pids and remove them if necessary
check_stale_pids(){
  check_status
  # If there is a pid it is something else than 0, the service is running if
  # brog_status is == 0.
  if [ "$wpid" != "0" -a "$brog_status" != "0" ]; then
    echo "Removing stale brog pid. This is most likely caused by brog crashing the last time it ran."
    if ! rm "$brog_pid_path"; then
      echo "Unable to remove stale pid, exiting"
      exit 1
    fi
  fi
}

# Check if a socket file was left behind and remove it if necessary
check_stale_socket(){
    # Only clean if brog not running
    if [ -e "$brog_socket_path" ]; then
    echo "Removing stale brog socket. This is most likely caused by brog crashing the last time it ran."
    echo "You may want to tell this to the devs. See https://github.com/aybabtme/brog/issue"
    if ! rm "$brog_socket_path"; then
      echo "Unable to remove stale socket, exiting"
      exit 1
    fi
  fi
}

# If no parts of the service is running, bail out.
exit_if_not_running(){
  check_stale_pids
  if [ "$brog_status" != "0" ]; then
    echo "Brog is not running."
    exit
  fi
}

# Starts Brog.
start() {
  check_stale_pids
  check_stale_socket

  # Then check if the service is running. If it is: don't start again.
  if [ "$brog_status" = "0" ]; then
    echo "The Brog instance is already running with pid '$wpid', not restarting."
  else
    echo "Starting the Brog instance..."
    
    ./brog server prod & disown
  fi
  # Let settle because in some configuration the following will report brog is not running because it's still starting
  sleep 1  
  # Finally check the status to tell whether or not brog is running
  status
}

# Asks brog if it would be so kind as to stop, if not kills it.
stop() {
  exit_if_not_running
  # If the Brog instance is running, tell it to stop;
  if [ "$brog_status" = "0" ]; then
    # Send SIGINT to terminate process gracefully
    kill -2 $wpid
    echo "Stopping the Brog instance..."
    stopping=true
  else
    echo "The Brog instance was not running, doing nothing."
  fi

  # If something needs to be stopped, lets wait for it to stop. Never use SIGKILL in a script.
  while [ "$stopping" = "true" ]; do
    sleep 1
    check_status
    if [ "$brog_status" = "0"  ]; then
      printf "."
    else
      printf "\n"
      break
    fi
  done
  sleep 1
  # Cleaning up unused pids
  if [ -e "$brog_pid_path" ]; then
    echo "Removing stale brog pidfile. This is most likely caused by brog not cleaning it."
    echo "You may want to investigate or tell this to the devs. See https://github.com/aybabtme/brog/issue"
    if ! rm "$brog_pid_path"; then
      echo "Unable to remove stale pid!!"
    fi
  fi
  status
}

# Returns the status of Brog and it's components
status() {
  check_status
  if [ "$brog_status" != "0" ]; then
    echo "Brog is not running."
    return
  fi
  if [ "$brog_status" = "0" ]; then
      echo "The Brog instance with pid '$wpid' is running."
  else
      printf "The Brog instance is \033[31mnot running\033[0m.\n"
  fi
}

restart(){
  check_status
  if [ "$brog_status" = "0" ]; then
    stop
  fi
  start
}

## Finally the input handling.

case "$1" in
  start)
        start
        ;;
  stop)
        stop
        ;;
  restart)
        restart
        ;;
  status)
        status
        ;;
  *)
        echo "Usage: service brog {start|stop|restart|status}"
        exit 1
        ;;
esac

exit
