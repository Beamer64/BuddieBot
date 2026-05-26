//go:build windows

package lavalink_runner

import (
	"fmt"
	"os/exec"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// A Job Object with KILL_ON_JOB_CLOSE ensures the child dies with us on
// any exit path — handle release at process exit fires the kill trigger.

var (
	jobOnce   sync.Once
	jobHandle windows.Handle
	jobErr    error
)

func ensureJob() (windows.Handle, error) {
	jobOnce.Do(func() {
		h, err := windows.CreateJobObject(nil, nil)
		if err != nil {
			jobErr = fmt.Errorf("create job object: %w", err)
			return
		}
		var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
		info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
		if _, err := windows.SetInformationJobObject(
			h,
			windows.JobObjectExtendedLimitInformation,
			uintptr(unsafe.Pointer(&info)),
			uint32(unsafe.Sizeof(info)),
		); err != nil {
			_ = windows.CloseHandle(h)
			jobErr = fmt.Errorf("set kill-on-close on job: %w", err)
			return
		}
		jobHandle = h
	})
	return jobHandle, jobErr
}

func prepareCmd(cmd *exec.Cmd) error {
	_, err := ensureJob()
	return err
}

func attachReaper(cmd *exec.Cmd) error {
	job, err := ensureJob()
	if err != nil {
		return err
	}
	procH, err := windows.OpenProcess(
		windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE,
		false,
		uint32(cmd.Process.Pid),
	)
	if err != nil {
		return fmt.Errorf("open child process: %w", err)
	}
	defer windows.CloseHandle(procH)
	if err := windows.AssignProcessToJobObject(job, procH); err != nil {
		return fmt.Errorf("assign to job: %w", err)
	}
	return nil
}
