import os
import subprocess
import sys

import snapcraft

class XGobuildPlugin(snapcraft.BasePlugin):
    def build(self):
        env = os.environ.copy()

        # All executables in a classic snap must be linked with special flags.
        # Unfortunately, Go's ./make.bash does not respect the LDFLAGS
        # environment variable and so we use a wrapper which does and tell Go to
        # use that as the linker.
        mycc = os.path.join(self.builddir, 'mycc')
        with open(mycc, 'w') as script:
            os.chmod(script.fileno(), 0o755)
            script.write('#!/bin/bash\n')
            script.write('set -ex\n')
            script.write('exec gcc $LDFLAGS "$@"\n')
        env['GO_LDFLAGS'] = '-linkmode=external -extld=%s'%(mycc,)

        # Bootstrap with the go that is on the PATH.
        goroot_bootstrap = subprocess.check_output(['go', 'env', 'GOROOT'])
        env['GOROOT_BOOTSTRAP'] = goroot_bootstrap.decode(sys.getfilesystemencoding()).rstrip('\n')

        self.run(['./make.bash'], cwd=os.path.join(self.builddir, 'src'), env=env)

        # Remove our gcc wrapper.
        os.unlink(mycc)

        # Just ship the whole tree.
        self.run(['rsync', '-a', '--exclude', '.git', self.builddir + '/', self.installdir])

        # And finally, create a wrapper that sets $GOROOT based on $SNAP.
        with open(os.path.join(self.installdir, 'gowrapper'), 'w') as script:
            os.chmod(script.fileno(), 0o755)
            script.write('#!/bin/bash\n')
            script.write('export GOROOT="$SNAP"\n')
            script.write('exec $SNAP/bin/go "$@"\n')
