require 'spec_helper'

describe "Authentication" do
  subject { page }

  describe "signin page" do
    before { visit signin_path }

    it { should have_content("Sign in") }
    it { should have_title("Sign in") }
  end

  describe "sign in" do
    before { visit signin_path }

    describe "with invalid information" do
      before { click_button("Sign in") }

      it { should have_title("Sign in") }
      it { should have_error_message("Invalid") }

      describe "after visiting another page" do
        before { click_link "Home" }
        it { should_not have_selector("div.alert.alert-error") }
      end
    end

    describe "with valid information" do
      let(:user) { FactoryGirl.create(:user) }
      before { sign_in(user) }

      it { should have_title(user.name) }
      it { should have_link("Users", href: users_path) }
      it { should have_link("Profile", href: user_path(user)) }
      it { should have_link("Settings", href: edit_user_path(user)) }
      it { should have_link("Sign out", href: signout_path) }
      it { should_not have_link("Sign in", href: signin_path) }

      describe "followed by signout" do
        before { click_link "Sign out" }
        it { should have_link("Sign in") }
      end
    end
  end

  describe "authorization" do
    describe "for signed-in userd" do
      let(:user) { FactoryGirl.create(:user) }
      before { sign_in user, no_capybara: true }

      describe "in the Users controller" do
        describe "visiting the new action" do
          before { get new_user_path }
          specify { expect(response).to redirect_to(root_path) }
        end

        describe "submitting to the create action" do
          before { post users_path }
          specify { expect(response).to redirect_to(root_path) }
        end
      end
    end

    describe "for non-signed-in users" do
      let(:user) { FactoryGirl.create(:user) }

      describe "in the Users controller" do
        before { visit edit_user_path(user) }
        it { should have_title("Sign in") }
        it { should_not have_link("Profile") }
        it { should_not have_link("Settings") }

        describe "visiting the user indes" do
          before { visit users_path }
          it { should have_title("Sign in") }
        end
      end

      describe "submitting to the update action" do
        before { patch user_path(user) }
        specify { expect(response).to redirect_to(signin_path) }
      end

      describe "when attempting to visit a protected page" do
        before do
          visit edit_user_path(user)
          fill_in "Email", with: user.email
          fill_in "Password", with: user.password
          click_button "Sign in"
        end

        describe "after signing in" do
          it "should render the desired protected page" do
            expect(page).to have_title("Edit user")
          end

          describe "when signing in again" do
            before do
              delete signout_path
              visit signin_path
              fill_in "Email", with: user.email
              fill_in "Password", with: user.password
              click_button "Sign in"
            end

            it "should render the default (profile) page" do
              expect(page).to have_title(user.name)
            end
          end
        end
      end

      describe "as non-admin user" do
        let(:user) { FactoryGirl.create(:user) }
        let(:non_admin) { FactoryGirl.create(:user) }

        before { sign_in non_admin, no_capybara: true }

        describe "submitting a DELETE request to the Users#destroy action" do
          before { delete user_path(user) }
          specify { expect(response).to redirect_to(root_path) }
        end
      end
    end

    describe "as wrong user" do
      let(:user) { FactoryGirl.create(:user) }
      let(:wrong_user) { FactoryGirl.create(:user, email: "wrong@example.com") }
      before { sign_in user, no_capybara: true }

      describe "submitting a GET request to the Users#edit action" do
        before { get edit_user_path(wrong_user) }
        specify { expect(response.body).not_to match(full_title("Edit user")) }
        specify { expect(response).to redirect_to(root_url) }
      end

      describe "submitting a PATCH request to the Users#update action" do
        before { patch user_path(wrong_user) }
        specify { expect(response).to redirect_to(root_path) }
      end
    end

    describe "as an admin user" do
      let(:admin) { FactoryGirl.create(:admin) }
      before { sign_in admin, no_capybara: true }

      describe "submitting a DELETE request to the Users#destroy action" do
        before { delete user_path(admin) }
        specify { expect(response).to redirect_to(root_path) }
      end
    end
  end
end
