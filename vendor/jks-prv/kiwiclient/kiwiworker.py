## -*- python -*-

import logging
import threading
from traceback import print_exc

from kiwiclient import KiwiTooBusyError
from kiwiclient import KiwiTimeLimitError
from kiwiclient import KiwiServerTerminatedConnection

class KiwiWorker(threading.Thread):
    def __init__(self, group=None, target=None, name=None, args=(), kwargs=None):
        super(KiwiWorker, self).__init__(group=group, target=target, name=name)
        self._recorder, self._options, self._run_event = args
        self._recorder._reader = True
        self._event = threading.Event()

    def _do_run(self):
        return self._run_event.is_set()

    def run(self):
        while self._do_run():
            try:
                self._recorder.connect(self._options.server_host, self._options.server_port)
            except Exception as e:
                logging.info("Failed to connect, sleeping and reconnecting error='%s'" %e)
                if self._options.is_kiwi_tdoa:
                    self._options.status = 1
                    break
                self._event.wait(timeout=15)
                continue

            try:
                self._recorder.open()
                while self._do_run():
                    self._recorder.run()
            except KiwiServerTerminatedConnection as e:
                logging.info("%s:%d %s. Reconnecting after 5 seconds"
                      % (self._options.server_host, self._options.server_port, e))
                self._recorder.close()
                self._recorder._start_ts = None ## this makes the recorder to open a new file on restart
                self._event.wait(timeout=5)
                continue
            except KiwiTooBusyError:
                logging.info("%s:%d too busy now. Reconnecting after 15 seconds"
                      % (self._options.server_host, self._options.server_port))
                if self._options.is_kiwi_tdoa:
                    self._options.status = 2
                    break
                self._event.wait(timeout=15)
                continue
            except KiwiTimeLimitError:
                break
            except Exception as e:
                if self._options.is_kiwi_tdoa:
                    self._options.status = 1
                print_exc()
                break

        self._run_event.clear()   # tell all other threads to stop
        self._recorder.close()
