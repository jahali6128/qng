// Copyright 2012 Jimmy Zelinskie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package whirlpool_test

import (
	"fmt"
	"github.com/Qitmeer/qng/crypto/x16rv3/whirlpool"
	"io"
	"testing"
)

type whirlpoolTest struct {
	out string
	in  string
}

var golden = []whirlpoolTest{
	{"19FA61D75522A4669B44E39C1D2E1726C530232130D407F89AFEE0964997F7A73E83BE698B288FEBCF88E3E03C4F0757EA8964E59B63D93708B138CC42A66EB3", ""},
	{"8ACA2602792AEC6F11A67206531FB7D7F0DFF59413145E6973C45001D0087B42D11BC645413AEFF63A42391A39145A591A92200D560195E53B478584FDAE231A", "a"},
	{"33E24E6CBEBF168016942DF8A7174048F9CEBC45CBD829C3B94B401A498ACB11C5ABCCA7F2A1238AAF534371E87A4E4B19758965D5A35A7CAD87CF5517043D97", "ab"},
	{"4E2448A4C6F486BB16B6562C73B4020BF3043E3A731BCE721AE1B303D97E6D4C7181EEBDB6C57E277D0E34957114CBD6C797FC9D95D8B582D225292076D4EEF5", "abc"},
	{"BDA164F0B930C43A1BACB5DF880B205D15AC847ADD35145BF25D991AE74F0B72B1AC794F8AACDA5FCB3C47038C954742B1857B5856519DE4D1E54BFA2FA4EAC5", "abcd"},
	{"5D745E26CCB20FE655D39C9E7F69455758FBAE541CB892B3581E4869244AB35B4FD6078F5D28B1F1A217452A67D9801033D92724A221255A5E377FE9E9E5F0B2", "abcde"},
	{"A73E425459567308BA5F9EB2AE23570D0D0575EB1357ECF6AC88D4E0358B0AC3EA2371261F5D4C070211784B525911B9EEC0AD968429BB7C7891D341CFF4E811", "abcdef"},
	{"08B388F68FD3EB51906AC3D3C699B8E9C3AC65D7CEB49D2E34F8A482CBC3082BC401CEAD90E85A97B8647C948BF35E448740B79659F3BEE42145F0BD653D1F25", "abcdefg"},
	{"1F1A84D30612820243AFE2022712F9DAC6D07C4C8BB41B40EACAB0184C8D82275DA5BCADBB35C7CA1960FF21C90ACBAE8C14E48D9309E4819027900E882C7AD9", "abcdefgh"},
	{"11882BC9A31AC1CF1C41DCD9FD6FDD3CCDB9B017FC7F4582680134F314D7BB49AF4C71F5A920BC0A6A3C1FF9A00021BF361D9867FE636B0BC1DA1552E4237DE4", "abcdefghi"},
	{"717163DE24809FFCF7FF6D5ABA72B8D67C2129721953C252A4DDFB107614BE857CBD76A9D5927DE14633D6BDC9DDF335160B919DB5C6F12CB2E6549181912EEF", "abcdefghij"},
	{"2C06DA809D8497667DE1563A2AC1C6D8DF8233D7C1E6CCB2E3DA542BD237DF553AA90AD0DDF3AEFB711FBBD26C36F667408206DDC8047736987075805803A315", "Discard medicine more than two years old."},
	{"510B9442AFED7E7945D7524D00AF0239E84EB0B3644EFA9481EE62154D04D82680BEC1741701AAF98D8C887BB875F15399CC11CF27A27E066B2ADCF7E443CEB2", "He who has a shady past knows that nice guys finish last."},
	{"761D7DB6292384CCC4A806A18404031D89DBBCE5C22BB284A1E5D5979F44E37348857E555BABF61B7EACBDC8DF543F6477A5611330866D6660ED7C62655A5555", "I wouldn't marry him with a ten foot pole."},
	{"C87BD0FEAA7146BD0576796CE48ABC05D861C31C567599E4A8DD01B9D66536E373229F7A0BFA32FBD438CD64C56CA405DBF0C83CD89C9F8F9A750ACFFF9B255D", "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
	{"B30647F490036F0AD029E6E882296A87757B163679CC910A0B9DC5D26C1572B3467C8F198EED4AAF8A592041A38A673A53BCAB1CF87647AEC2119A1875A149A8", "The days of the digital watch are numbered.  -Tom Stoppard"},
	{"E37E7BE075772E277EA7DF46E317B13E7B748B12BC214F7A55D9ED230C13C73FECB573A0AC216F2F59C15E32609786263D933CAD9E8C8009293EBD42A7626672", "Nepal premier won't resign."},
	{"96E52A21BDC3B1486CDBCE43059DF4D765290B6D330C4B01A68A8081930FB0DEF854EA6260A022CC26D8257262BC71B4B0DB1CBA546F849D3839738BE56ADD04", "For every action there is an equal and opposite government program."},
	{"636CE4B33A674AAF0BB4E2371D85B116019D5708DC6434F6DB9A455151981F2C36D7C500B040AF09EC2D5A394018EB656AE21356564318A0FE8F0D1C1B6750EC", "His money is twice tainted: 'taint yours and 'taint mine."},
	{"41C151D2FE36CEA72F51729CA38D255EE062F34A5303C7445E8113B970E24B4AF52E05B74CB5E7EBE06E26EAE8CC1476736012D92E5582750F86123BD378F48D", "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
	{"4111F7CD1B917CC403B120A848C2F5F3B6FF4324190438FDE65EC74C16105340C0852400200D99FADF51910A15F3BAC49C1C1953745082FAF992A2AE24C71086", "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
	{"7760D96148F6AE372E3C8A499654D7EF0CB7C1DCBEB441FF9DC7576416C39C82490D5162AEB595C207D16C3850F2C6F39EA8DFBD865EBD8EABFA2A0FCDBBA245", "size:  a.out:  bad magic"},
	{"9FE97F7F0AC799D710F6EDA3905D89E5267FA74E87F57174473AA7ADE31BF07DDC0448908E14C9E65F7814661E3596501D3E50D51F46CAFEC661942BDA922D79", "The major problem is with sendmail.  -Mark Horton"},
	{"225F9AE6AFC27412F8AED205E1A91996CDB8C75988E4CF3783E8A0D0B62C0FDDC4BA6133F262608BF0EABE188C99A497F9E948AEFD1F1558AA322BFF9967C96F", "Give me a rock, paper and scissors and I will move the world.  CCFestoon"},
	{"954ADA35E1D0A0A909131B7CEA6140B9BF3336E83420336FC292B6E9BF5BC724AFDFFDB0F4316D221FD628C51EC81B36466288F95C20DC223874BC99450EFA6F", "If the enemy is within range, then so are you."},
	{"1B3E60898B4CFDCBB691EF4EAB151E03BC48F45823726D8035EFC21E35EB5E635079D2F6BC813D0B850B7B6B18D8BC1580225392A4EE0D4B472058895D34189F", "It's well we cannot hear the screams/That we create in others' dreams."},
	{"06F11A0193C8F65458123D756A0F0C81D33A5716540F8DDEA2D6EDAFAA426111428A83F251FD2111E42AFD4F2854B4A58157DC714730892AD10C7338EA28BD1C", "You remind me of a TV show, but that's all right: I watch it anyway."},
	{"7650550B8B5CB9077D92AAEECA7893B5BF5F1CBE65AB1461F5797B4B1C4141F0F6A6F9F4854EBEBF77651F117892F1C8A9E48C7A6E27A589DDF6EC46255883D5", "C is as portable as Stonehedge!!"},
	{"4DA620619C79436319EAB7CBFE59C2E248264837528378620DB9D4B0B5BA2F79532156CFE37A8D16F455AAD7FABDE46B48DEF2BE7CB564389A0FD341B4F3E57E", "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
	{"D405124ABE5C46AEBD4FB71BF7C3C56AD0E0F25728789EBFFB68648BFD8786765194CA0BF6D0EF11AE0E5B9381A00E5B04E80C8FE98F7EFF65ED3655113A4A06", "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction.  Lewis-Randall Rule"},
	{"863E4A9320C80173A332D84C495B818104650B1FEAF39A9E6438415C33EB2F2F41630655B7A72E798DE0A9E12FBEF367B2FE241AC93CDFE02FE8F3EEC8C518F4", "How can you write a big system without C++?  -Paul Glick"},
}

func TestGolden(t *testing.T) {
	for i := 0; i < len(golden); i++ {
		g := golden[i]
		c := whirlpool.New()
		for j := 0; j < 3; j++ {
			if j < 2 {
				io.WriteString(c, g.in)
			} else {
				io.WriteString(c, g.in[0:len(g.in)/2])
				c.Sum(nil)
				io.WriteString(c, g.in[len(g.in)/2:])
			}
			s := fmt.Sprintf("%X", c.Sum(nil))
			if s != g.out {
				t.Fatalf("whirlpool[%d](%s) = %s want %s", j, g.in, s, g.out)
			}
			c.Reset()
		}
	}
}

func ExampleNew() {
	h := whirlpool.New()
	io.WriteString(h, "His money is twice tainted: 'taint yours and 'taint mine.")
	fmt.Printf("% x", h.Sum(nil))
	// Output:
	// 63 6c e4 b3 3a 67 4a af 0b b4 e2 37 1d 85 b1 16 01 9d 57 08 dc 64 34 f6 db 9a 45 51 51 98 1f 2c 36 d7 c5 00 b0 40 af 09 ec 2d 5a 39 40 18 eb 65 6a e2 13 56 56 43 18 a0 fe 8f 0d 1c 1b 67 50 ec
}
