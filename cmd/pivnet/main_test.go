package main_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

const (
	apiPrefix = "/api/v2"
	apiToken  = "some-api-token"
)

var _ = Describe("pivnet cli", func() {
	var (
		server   *ghttp.Server
		endpoint string
		args     []string

		product  pivnet.Product
		products []pivnet.Product

		eulas []pivnet.EULA

		releases []pivnet.Release
		release  pivnet.Release
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		endpoint = server.URL()

		args = []string{}

		product = pivnet.Product{
			ID:   1234,
			Slug: "some-product-slug",
			Name: "some-product-name",
		}

		products = []pivnet.Product{
			product,
			{
				ID:   2345,
				Slug: "another-product-slug",
				Name: "another product name",
			},
		}

		eulas = []pivnet.EULA{
			{
				ID:   1234,
				Name: "some eula",
				Slug: "some-eula",
			},
			{
				ID:   2345,
				Name: "another eula",
				Slug: "another-eula",
			},
		}

		release = pivnet.Release{
			ID:      1234,
			Version: "version 0.2.3",
		}

		releases = []pivnet.Release{
			release,
			{
				ID:      2345,
				Version: "version 0.3.4",
			},
		}

	})

	runMainWithArgs := func(args ...string) *gexec.Session {
		args = append(
			args,
			fmt.Sprintf("--api-token=%s", apiToken),
			fmt.Sprintf("--endpoint=%s", endpoint),
		)

		_, err := fmt.Fprintf(GinkgoWriter, "Running command: %v\n", args)
		Expect(err).NotTo(HaveOccurred())

		command := exec.Command(pivnetBinPath, args...)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

	Describe("Displaying help", func() {
		It("displays help with '-h'", func() {
			session := runMainWithArgs("-h")

			Eventually(session, executableTimeout).Should(gexec.Exit())
			Expect(session.Err).Should(gbytes.Say("Usage"))
		})

		It("displays help with '--help'", func() {
			session := runMainWithArgs("--help")

			Eventually(session, executableTimeout).Should(gexec.Exit())
			Expect(session.Err).Should(gbytes.Say("Usage"))
		})
	})

	Describe("Displaying version", func() {
		It("displays version with '-v'", func() {
			session := runMainWithArgs("-v")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say("dev"))
		})

		It("displays version with '--version'", func() {
			session := runMainWithArgs("--version")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say("dev"))
		})
	})

	Describe("printing as json", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s", apiPrefix, product.Slug),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, product),
				),
			)
		})

		It("prints as json", func() {
			session := runMainWithArgs("--print-as=json", "product", "-s", product.Slug)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			var receivedProduct pivnet.Product
			err := json.Unmarshal(session.Out.Contents(), &receivedProduct)
			Expect(err).NotTo(HaveOccurred())

			Expect(receivedProduct.Slug).To(Equal(product.Slug))
		})
	})

	Describe("printing as yaml", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s", apiPrefix, product.Slug),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, product),
				),
			)
		})

		It("prints as yaml", func() {
			session := runMainWithArgs("--print-as=yaml", "product", "-s", product.Slug)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			var receivedProduct pivnet.Product
			err := yaml.Unmarshal(session.Out.Contents(), &receivedProduct)
			Expect(err).NotTo(HaveOccurred())

			Expect(receivedProduct.Slug).To(Equal(product.Slug))
		})
	})

	Describe("product", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s", apiPrefix, product.Slug),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, product),
				),
			)
		})

		It("displays product for the provided slug", func() {
			session := runMainWithArgs("product", "-s", product.Slug)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(product.Slug))
		})
	})

	Describe("EULAs", func() {
		It("displays all EULAs", func() {
			eulasResponse := pivnet.EULAsResponse{
				EULAs: eulas,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas", apiPrefix)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, eulasResponse),
				),
			)

			session := runMainWithArgs("eulas")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(eulas[0].Name))
			Expect(session).Should(gbytes.Say(eulas[1].Name))
		})

		It("accepts EULAs", func() {
			eulaAcceptanceResponse := pivnet.EULAAcceptanceResponse{
				AcceptedAt: "now",
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"POST",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/eula_acceptance",
							apiPrefix,
							product.Slug,
							release.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, eulaAcceptanceResponse),
				),
			)

			session := runMainWithArgs(
				"accept-eula",
				"--product-slug", product.Slug,
				"--release-id", strconv.Itoa(release.ID),
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
		})
	})

	Describe("Release Types", func() {
		var (
			releaseTypes []string
		)

		BeforeEach(func() {
			releaseTypes = []string{"some release type", "another release type"}
		})

		It("displays all Release Types", func() {
			releaseTypesResponse := pivnet.ReleaseTypesResponse{
				ReleaseTypes: releaseTypes,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/releases/release_types", apiPrefix)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releaseTypesResponse),
				),
			)

			session := runMainWithArgs("release-types")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(releaseTypes[0]))
			Expect(session).Should(gbytes.Say(releaseTypes[1]))
		})
	})
})
