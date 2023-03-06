package goproc

import (
	"testing"
)

func TestGetEnvironFromPsCommand(t *testing.T) {
	cases := []struct {
		str    string
		except int
		msg    string
	}{
		{"command SHELL=/bin/zsh", 1 , "環境変数1つ"},
		{"command TERM=xterm-256color SHELL=/bin/zsh", 2 , "一般的な環境変数"},
		{"command LESS=-F -g -i -M -R -S -w -X -z-4 TERM=xterm-256color SHELL=/bin/zsh", 3, "スペース入りの環境変数（前）"},
		{"command PERL_LOCAL_LIB_ROOT=/Users/gozu/perl5 LESS=-F -g -i -M -R -S -w -X -z-4 TERM=xterm-256color", 3, "スペース入りの環境変数（中）"},
		{"command ITERM_SESSION_ID=w0t3p2:967773C8-DBEA-4BB9-8F87-C8A89EAF26A4 PERL_LOCAL_LIB_ROOT=/Users/gozu/perl5 LESS=-F -g -i -M -R -S -w -X -z-4", 3, "スペース入りの環境変数（後）"},
		{"command ERL_LOCAL_LIB_ROOT=/Users/gozu/perl5 PERL_MB_OPT=--install_base \"/Users/gozu/perl5\"", 2, "スペースが入った環境変数(ダブルコーテーション付き)後"},
		{"command PERL_MB_OPT=--install_base \"/Users/gozu/perl5\" PERL_MM_OPT=INSTALL_BASE=/Users/gozu/perl5", 2, "スペースが入った環境変数(ダブルコーテーション付き)前"},
		{"command USER=gozu LS_COLORS=di=34:ln=35:so=32:pi=33:ex=31:bd=36;01:cd=33;01:su=31;40;07:sg=36;40;07:tw=32;40;07:ow=33;40;07: COMMAND_MODE=unix2003 GREP_COLORS=mt=37;45 /Users/gozu/INFOCOM/ism=", 5, "=が入った環境変数"},
		{"command PERL_MB_OPT=--install_base \"/Users/gozu/perl5\" PERL_MM_OPT=INSTALL_BASE=/Users/gozu/perl5", 2, "スペースと=が入った環境変数"},
		{"command COMMAND_MODE=unix2003 GREP_COLORS=mt=37;45 /Users/gozu/INFOCOM/ism=", 3, "Valueがない環境変数"},
		{"/Users/gozu/projects/nemu/nemu_mc2/node_modules/esbuild-darwin-64/bin/esbuild --service=0.15.12 --ping USER=gozu LS_COLORS=di=34:ln=35:so=32:pi=33:ex=31:bd=36;01:cd=33;01:su=31;40;07:sg=36;40;07:tw=32;40;07:ow=33;40;07: COMMAND_MODE=unix2003 GREP_COLORS=mt=37;45 /Users/gozu/INFOCOM/ism=", 5, "commandを除去するテスト"},
		{"/Applications/MacVim.app/Contents/MacOS/Vim ismService_darwin.go USER=gozu LS_COLORS=di=34:ln=35:so=32:pi=33:ex=31:bd=36;01:cd=33;01:su=31;40;07:sg=36;40;07:tw=32;40;07:ow=33;40;07: COMMAND_MODE=unix2003 GREP_COLORS=mt=37;45 /Users/gozu/INFOCOM/ism=", 5, "commandを除去するテスト2"},
		{"/usr/bin/java -Xms64M -Xmx1G -Djava.util.logging.config.file=logging.properties -Djava.security.auth.login.config=/Users/gozu/INFOCOM/ism/service/activemq//conf/login.config -Dcom.sun.management.jmxremote -Djava.awt.headless=true -Djava.io.tmpdir=/Users/gozu/INFOCOM/ism/service/activemq//tmp --add-reads=java.xml=java.logging --add-opens java.base/java.security=ALL-UNNAMED --add-opens java.base/java.net=ALL-UNNAMED --add-opens java.base/java.lang=ALL-UNNAMED --add-opens java.base/java.util=ALL-UNNAMED --add-opens java.naming/javax.naming.spi=ALL-UNNAMED --add-opens java.rmi/sun.rmi.transport.tcp=ALL-UNNAMED --add-opens java.base/java.util.concurrent=ALL-UNNAMED --add-opens java.base/java.util.concurrent.atomic=ALL-UNNAMED --add-exports=java.base/sun.net.www.protocol.http=ALL-UNNAMED --add-exports=java.base/sun.net.www.protocol.https=ALL-UNNAMED --add-exports=java.base/sun.net.www.protocol.jar=ALL-UNNAMED --add-exports=jdk.xml.dom/org.w3c.dom.html=ALL-UNNAMED --add-exports=jdk.naming.rmi/com.sun.jndi.url.rmi=ALL-UNNAMED -Dactivemq.classpath=/Users/gozu/INFOCOM/ism/service/activemq//conf:/Users/gozu/INFOCOM/ism/service/activemq//../lib/: -Dactivemq.home=/Users/gozu/INFOCOM/ism/service/activemq/ -Dactivemq.base=/Users/gozu/INFOCOM/ism/service/activemq/ -Dactivemq.conf=/Users/gozu/INFOCOM/ism/service/activemq//conf -Dactivemq.data=/Users/gozu/INFOCOM/ism/service/activemq//data -jar /Users/gozu/INFOCOM/ism/service/activemq//bin/activemq.jar start /Users/gozu/INFOCOM/ism= SHELL=/bin/zsh LSCOLORS=exfxcxdxbxGxDxabagacad ITERM_PROFILE=desktop2", 4, "java系commandを除去するテスト"},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			result := getEnvironFromPsCommand(c.str)
			if len(result) != c.except {
				t.Errorf("getEnvironFromPsCommand = %d, expect = %d, Failed", len(result), c.except)
			}
		})
	}
}
